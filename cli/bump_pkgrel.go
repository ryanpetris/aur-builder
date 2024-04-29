package cli

import (
	"flag"
	"fmt"
	"github.com/ryanpetris/aur-builder/cienv"
	"github.com/ryanpetris/aur-builder/git"
	"github.com/ryanpetris/aur-builder/pkg"
	"log/slog"
	"slices"
	"strings"
)

func BumpPkgrel(args []string) {
	cmd := flag.NewFlagSet("bump-pkgrel", flag.ExitOnError)

	cmdPackages := cmd.String("packages", "", "comma-separated list of packages to bump")

	if err := cmd.Parse(args[1:]); err != nil {
		panic(err)
	}

	packages := strings.Split(*cmdPackages, ",")

	if len(packages) == 0 {
		panic("--packages is required")
	}

	cenv := cienv.FindCiEnv()
	var bumpPkgbase []string

	allPackages, err := pkg.GetPackages()

	if err != nil {
		panic(err)
	}

	for _, pkgbase := range allPackages {
		pconfig, err := pkg.LoadConfig(pkgbase)

		if err != nil {
			panic(err)
		}

		if pconfig.Ignore {
			continue
		}

		if err := pconfig.Merge(pkgbase, true); err != nil {
			panic(err)
		}

		pkgnames, err := pkg.GetMergedPkgnames(pkgbase)

		if err != nil {
			panic(err)
		}

		for _, pkgname := range pkgnames {
			if !slices.Contains(packages, pkgname) {
				continue
			}

			bumpPkgbase = append(bumpPkgbase, pkgbase)
			break
		}
	}

	for _, pkgbase := range bumpPkgbase {
		pconfig, err := pkg.LoadConfig(pkgbase)

		if err != nil {
			panic(err)
		}

		if err := pconfig.Merge(pkgbase, true); err != nil {
			panic(err)
		}

		upstreamEpoch, mergedPkgver, mergedPkgrel, mergedSubpkgrel, err := pkg.GetMergedVersionParts(pkgbase)

		if err != nil {
			panic(err)
		}

		var branchVersion string

		if pconfig.Vcs != nil {
			pconfig.Vcs.Pkgrel += 1
			branchVersion = pkg.GetVersionString(upstreamEpoch, pconfig.Vcs.Pkgver, pconfig.Vcs.Pkgrel, 0)
		} else {
			if pconfig.Overrides == nil {
				pconfig.Overrides = &pkg.PackageConfigOverrides{}
			}

			if pconfig.Overrides.BumpPkgrel == nil {
				pconfig.Overrides.BumpPkgrel = map[string]int{}
			}

			pconfig.Overrides.BumpPkgrel[mergedPkgver] += 1
			branchVersion = pkg.GetVersionString(upstreamEpoch, mergedPkgver, mergedPkgrel+1, mergedSubpkgrel)
		}

		if exists, err := git.PackageUpdateBranchExists(pkgbase, branchVersion); err != nil {
			panic(err)
		} else if exists {
			slog.Info(fmt.Sprintf("Already have branch for updating pacakge %s to version %s. Skipping.", pkgbase, branchVersion))
			continue
		}

		slog.Info(fmt.Sprintf("Updating package %s", pkgbase))

		if cenv.IsCI() {
			if err := git.CreateAndSwitchToPackageUpdateBranch(pkgbase, branchVersion); err != nil {
				panic(err)
			}
		}

		if err := pconfig.Write(pkgbase); err != nil {
			panic(err)
		}

		if cenv.IsCI() {
			if err := git.AddAll(); err != nil {
				panic(err)
			}

			if err := git.Commit(fmt.Sprintf("Update %s at version %s", pkgbase, branchVersion)); err != nil {
				panic(err)
			}

			if err := git.PushPackageBranch(pkgbase, branchVersion); err != nil {
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
