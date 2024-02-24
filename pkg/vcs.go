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

			if source.FragmentType == "commit" {
				continue
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

	vcsPkgver, vcsPkgrel, err := GetMergedVcsPkgver(pkgbase)

	if err != nil {
		return false, err
	}

	vcinfo := &PackageVcs{}
	vcinfo.Pkgver = vcsPkgver
	vcinfo.Pkgrel = vcsPkgrel

	for _, srcPath := range dirs {
		if !srcPath.IsDir() {
			continue
		}

		source := vcsSources[srcPath.Name()]

		if source == nil {
			continue
		}

		revision, err := git.GetRevision(path.Join(sourcesPath, srcPath.Name()))

		if err != nil {
			return false, err
		}

		source.FragmentType = "commit"
		source.FragmentValue = revision

		override := PackageConfigOverrideFromTo{
			From: source.Original,
			To:   source.String(),
		}

		if override.From != override.To {
			vcinfo.SourceOverrides = append(vcinfo.SourceOverrides, override)
		}
	}

	if pconfig.Vcs == nil {
		pconfig.Vcs = vcinfo

		return true, nil
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
