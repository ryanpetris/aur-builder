package cli

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/arch"
	"github.com/ryanpetris/aur-builder/cienv"
	"github.com/ryanpetris/aur-builder/misc"
	"github.com/ryanpetris/aur-builder/pacman"
	"github.com/ryanpetris/aur-builder/pkg"
	"log/slog"
)

func NeedsBuildMain(args []string) {
	trackers := map[string]misc.PackageTracker{}
	allPackages, err := pkg.GetPackages()

	if err != nil {
		panic(err)
	}

	for _, pkgbase := range allPackages {
		pconfig, err := pkg.LoadConfig(pkgbase)

		if err != nil {
			panic(err)
		}

		if pconfig.Vcs != nil && pconfig.Vcs.Pkgver == "" {
			slog.Info(fmt.Sprintf("Skipping VCS package %s without VCS information. Run update-vcs.", pkgbase))
			continue
		}

		tracker := misc.PackageTracker{
			Pkgbase: pkgbase,
		}

		tracker.UpstreamVersion, err = pkg.GetMergedVersion(pkgbase)

		if err != nil {
			panic(err)
		}

		pkginfos, err := pacman.LoadSrcinfo(pkgbase)

		if err != nil {
			panic(err)
		}

		for _, pkginfo := range pkginfos {
			if tracker.RepositoryVersion == "" {
				tracker.RepositoryVersion, _ = arch.GetPackageVersion(pkginfo.Pkgname)
			}

			tracker.Packages = append(tracker.Packages, misc.PackageInfo{
				Pkgbase:     pkginfo.Pkgbase,
				Pkgname:     pkginfo.Pkgname,
				FullVersion: pkginfo.GetFullVersion(),
				BuildDeps:   pkginfo.GetAllBuildDepends(),
			})
		}

		tracker.NeedsUpdate, err = pacman.IsVersionNewer(tracker.RepositoryVersion, tracker.UpstreamVersion)

		if err != nil {
			continue
		}

		if tracker.NeedsUpdate {
			if tracker.RepositoryVersion == "" {
				slog.Info(fmt.Sprintf("Considering new package %s, version %s.", pkgbase, tracker.UpstreamVersion))
			} else {
				slog.Info(fmt.Sprintf("Considering package %s, version %s is newer than %s.", pkgbase, tracker.UpstreamVersion, tracker.RepositoryVersion))
			}
		}

		trackers[pkgbase] = tracker
	}

	var updatePackages []string

	for pkgbase, tracker := range trackers {
		if !tracker.NeedsUpdate {
			continue
		}

		skip := false

		for _, pkgitem := range tracker.Packages {
			for _, dep := range pkgitem.BuildDeps {
				if otracker, hasKey := trackers[dep]; hasKey && otracker.NeedsUpdate && otracker.Pkgbase != pkgbase {
					slog.Info(fmt.Sprintf("Skipping %s for this run due to dependencies.", pkgbase))
					skip = true
					break
				}
			}

			if skip {
				break
			}
		}

		if !skip {
			updatePackages = append(updatePackages, pkgbase)
		}
	}

	cenv := cienv.FindCiEnv()
	err = cenv.WriteBuildPackages(updatePackages)

	if err != nil {
		panic(err)
	}
}
