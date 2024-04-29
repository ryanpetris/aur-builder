package pkg

import (
	"errors"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"github.com/ryanpetris/aur-builder/git"
	"github.com/ryanpetris/aur-builder/pacman"
	"log/slog"
	"os"
	"path"
)

const (
	fragmentTypeCommit = "commit"
)

func (pconfig *PackageConfig) GenVcsInfo(pkgbase string) (bool, error) {
	slog.Debug(fmt.Sprintf("Generating VCS Package Information for %s", pkgbase))

	if err := pconfig.Merge(pkgbase, false); err != nil {
		return false, err
	}

	result, err := GetMergedSources(pkgbase)

	if err != nil {
		return false, err
	}

	vcsSources := map[string]*pacman.Source{}
	vcsSourceTypes := map[string]string{}

	for srcType, srcItems := range result {
		for _, srcItem := range srcItems {
			source, err := pacman.ParseSource(srcItem)

			if err != nil {
				return false, err
			}

			if source.VcsType != "git" {
				continue
			}

			vcsSources[source.GetFolder()] = source
			vcsSourceTypes[source.GetFolder()] = srcType
		}
	}
	if len(vcsSources) == 0 {
		return false, nil
	}

	if err := pacman.DownloadSources(pkgbase); err != nil {
		return false, err
	}

	sourcesPath := path.Join(config.GetMergedPath(pkgbase), "src")
	dirs, err := os.ReadDir(sourcesPath)

	if err != nil {
		return false, err
	}

	vcsPkgver, vcsPkgrel, vcsSubPkgrel, err := GetMergedVcsPkgver(pkgbase)

	if err != nil {
		return false, err
	}

	vcinfo := &PackageVcs{}
	vcinfo.Pkgver = vcsPkgver
	vcinfo.Pkgrel = vcsPkgrel

	if vcsSubPkgrel > 0 {
		vcinfo.Pkgrel++
	}

	if pconfig.Vcs != nil {
		vcinfo.Submodules = pconfig.Vcs.Submodules
	}

	if vcinfo.Submodules != nil {
		for _, srcPath := range dirs {
			if !srcPath.IsDir() {
				continue
			}

			source := vcsSources[srcPath.Name()]

			if source == nil {
				continue
			}

			submoduleMap := map[string]string{}

			for targetName, submoduleConfig := range vcinfo.Submodules {
				if submoduleConfig.Source == source.Folder {
					submoduleMap[submoduleConfig.Name] = targetName
				}
			}

			if len(submoduleMap) == 0 {
				continue
			}

			submodules, err := git.GetSubmodules(path.Join(sourcesPath, srcPath.Name()))

			if err != nil {
				return false, err
			}

			for sourceName, targetName := range submoduleMap {
				targetSource := vcsSources[targetName]

				if targetSource != nil && targetSource.FragmentType != fragmentTypeCommit {
					submodule := submodules[sourceName]

					if submodule != nil {
						targetSource.FragmentType = fragmentTypeCommit
						targetSource.FragmentValue = submodule.Hash
					}
				}
			}
		}
	}

	for _, srcPath := range dirs {
		if !srcPath.IsDir() {
			continue
		}

		source := vcsSources[srcPath.Name()]

		if source == nil {
			continue
		}

		if source.FragmentType != fragmentTypeCommit {
			revision, err := git.GetRevision(path.Join(sourcesPath, srcPath.Name()))

			if err != nil {
				return false, err
			}

			source.FragmentType = fragmentTypeCommit
			source.FragmentValue = revision
		}

		override := &PackageConfigOverrideFromTo{
			From: source.Original,
			To:   source.String(),
		}

		if override.From != override.To {
			vcinfo.SourceOverrides = append(vcinfo.SourceOverrides, override)
		}
	}

	if pconfig.Vcs == nil {
		if len(vcinfo.SourceOverrides) > 0 {
			pconfig.Vcs = vcinfo

			return true, nil
		} else {
			return false, nil
		}
	} else if !pconfig.Vcs.IsEqual(vcinfo) {
		if isNewer, err := pacman.IsVersionNewer(pconfig.Vcs.Pkgver, vcinfo.Pkgver); err != nil {
			return false, err
		} else if isNewer {
			pconfig.Vcs = vcinfo

			return true, nil
		} else if isNewer, err := pacman.IsVersionNewer(vcinfo.Pkgver, pconfig.Vcs.Pkgver); err != nil {
			return false, err
		} else if isNewer {
			return false, errors.New(fmt.Sprintf("old version %s is newer than new version %s", pconfig.Vcs.Pkgver, vcinfo.Pkgver))
		} else {
			vcinfo.Pkgrel = pconfig.Vcs.Pkgrel + 1
			pconfig.Vcs = vcinfo

			return true, nil
		}
	}

	return false, nil
}
