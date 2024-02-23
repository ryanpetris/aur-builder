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

	hasVcSources := false
	sources := map[string]*pacman.Source{}
	sourceTypes := map[string]string{}

	for srcType, srcItems := range result {
		for _, srcItem := range srcItems {
			source, err := pacman.ParseSource(srcItem)

			if err != nil {
				return false, err
			}

			if source.FragmentType == "commit" {
				continue
			}

			sources[source.GetFolder()] = source
			sourceTypes[source.GetFolder()] = srcType

			if source.VcsType == "git" {
				hasVcSources = true
			}
		}
	}

	if !hasVcSources {
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

	vcinfo := &PackageVersionControlInformation{}
	vcinfo.Pkgver = vcsPkgver
	vcinfo.Pkgrel = vcsPkgrel

	for _, srcPath := range dirs {
		if !srcPath.IsDir() {
			continue
		}

		source := sources[srcPath.Name()]

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

	updated := false

	if pconfig.VcInfo == nil {
		pconfig.VcInfo = vcinfo
		updated = true
	} else if !pconfig.VcInfo.IsEqual(vcinfo) {
		if isNewer, err := pacman.IsVersionNewer(pconfig.VcInfo.Pkgver, vcinfo.Pkgver); err != nil {
			return false, err
		} else if isNewer {
			pconfig.VcInfo = vcinfo
			updated = true
		} else if isNewer, err := pacman.IsVersionNewer(vcinfo.Pkgver, pconfig.VcInfo.Pkgver); err != nil {
			return false, err
		} else if isNewer {
			return false, errors.New(fmt.Sprintf("old version %s is newer than new version %s", pconfig.VcInfo.Pkgver, vcinfo.Pkgver))
		} else {
			vcinfo.Pkgrel = pconfig.VcInfo.Pkgrel + 1
			pconfig.VcInfo = vcinfo
			updated = true
		}
	}

	return updated, nil
}
