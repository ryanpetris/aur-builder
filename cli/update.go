package cli

import (
	"flag"
	"fmt"
	"github.com/ryanpetris/aur-builder/cienv"
	"github.com/ryanpetris/aur-builder/git"
	"github.com/ryanpetris/aur-builder/impenv"
	"github.com/ryanpetris/aur-builder/misc"
	"github.com/ryanpetris/aur-builder/pacman"
	"github.com/ryanpetris/aur-builder/pkg"
	"log/slog"
	"slices"
	"strings"
)

func UpdateMain(args []string) {
	cmd := flag.NewFlagSet("update", flag.ExitOnError)

	cmdSource := cmd.String("source", "", "package source (aur, arch)")

	if err := cmd.Parse(args[1:]); err != nil {
		panic(err)
	}

	if *cmdSource == "" {
		panic("--source is required")
	}

	source := strings.ToLower(*cmdSource)
	var ienv impenv.ImportEnv

	switch source {
	case "aur":
		ienv = impenv.AurImportEnv{}
	case "arch":
		ienv = impenv.ArchImportEnv{}
	default:
		panic(fmt.Sprintf("Invalid source: %s", *cmdSource))
	}

	cenv := cienv.FindCiEnv()
	var updatePkgbase []string
	var updatePkgname []string

	allPackages, err := pkg.GetPackages()

	if err != nil {
		panic(err)
	}

	for _, pkgbase := range allPackages {
		pconfig, err := pkg.LoadConfig(pkgbase)

		if err != nil {
			panic(err)
		}

		if pconfig.Source == source {
			updatePkgbase = append(updatePkgbase, pkgbase)

			pkgnames, err := pkg.GetUpstreamPkgnames(pkgbase)

			if err != nil {
				panic(err)
			}

			updatePkgname = append(updatePkgname, pkgnames[:]...)
		}
	}

	pkginfos, err := ienv.GetPackageInfo(updatePkgname)

	if err != nil {
		panic(err)
	}

	trackers := map[string]misc.PackageTracker{}
	var foundPackages []string

	for _, pkginfo := range pkginfos {
		tracker, hasKey := trackers[pkginfo.Pkgbase]

		if hasKey {
			tracker.Packages = append(tracker.Packages, pkginfo)
		} else {
			tracker = misc.PackageTracker{
				Pkgbase:           pkginfo.Pkgbase,
				RepositoryVersion: pkginfo.FullVersion,
				Packages:          []misc.PackageInfo{pkginfo},
			}

			tracker.UpstreamVersion, err = pkg.GetUpstreamVersion(pkginfo.Pkgbase)

			if err != nil {
				panic(err)
			}

			tracker.NeedsUpdate, err = pacman.IsVersionNewer(tracker.UpstreamVersion, tracker.RepositoryVersion)

			if err != nil {
				panic(err)
			}

			trackers[pkginfo.Pkgbase] = tracker
		}

		foundPackages = append(foundPackages, pkginfo.Pkgname)
	}

	for _, upkg := range updatePkgname {
		if !slices.Contains(foundPackages, upkg) {
			slog.Warn(fmt.Sprintf("Package %s no longer exists in the %s repository", upkg, source))
		}
	}

	for _, tracker := range trackers {
		if !tracker.NeedsUpdate {
			continue
		}

		if exists, err := git.PackageUpdateBranchExists(tracker.Pkgbase, tracker.RepositoryVersion); err != nil {
			panic(err)
		} else if exists {
			slog.Info(fmt.Sprintf("Already have branch for updating pacakge %s to version %s. Skipping.", tracker.Pkgbase, tracker.RepositoryVersion))
			continue
		}

		slog.Info(fmt.Sprintf("Updating package %s", tracker.Pkgbase))

		if cenv.IsCI() {
			if err := git.CreateAndSwitchToPackageUpdateBranch(tracker.Pkgbase, tracker.RepositoryVersion); err != nil {
				panic(err)
			}
		}

		if err := ienv.PackageImport(tracker.Pkgbase, tracker.RepositoryVersion); err != nil {
			panic(err)
		}

		pconfig, err := pkg.LoadConfig(tracker.Pkgbase)

		if err != nil {
			panic(err)
		}

		if err := pconfig.Merge(tracker.Pkgbase, false); err != nil {
			panic(err)
		}

		if updated, err := pconfig.GenVcsInfo(tracker.Pkgbase); err != nil {
			panic(err)
		} else if updated {
			if err := pconfig.Write(tracker.Pkgbase); err != nil {
				panic(err)
			}
		}

		if cenv.IsCI() {
			if err := git.AddAll(); err != nil {
				panic(err)
			}

			if err := git.Commit(fmt.Sprintf("Update %s at version %s", tracker.Pkgbase, tracker.RepositoryVersion)); err != nil {
				panic(err)
			}

			if err := git.PushPackageBranch(tracker.Pkgbase, tracker.RepositoryVersion); err != nil {
				panic(err)
			}

			if err := cenv.CreatePR(); err != nil {
				panic(err)
			}

			if err := git.SwitchToMaster(); err != nil {
				panic(err)
			}
		}
	}
}
