package cli

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/cienv"
	"github.com/ryanpetris/aur-builder/git"
	"github.com/ryanpetris/aur-builder/pkg"
	"log/slog"
)

func UpdateVcsMain(args []string) {
	cenv := cienv.FindCiEnv()
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

		if pconfig.VcInfo == nil {
			continue
		}

		updated, err := pconfig.GenVcsInfo(pkgbase)

		if err != nil {
			panic(err)
		}

		if !updated {
			continue
		}

		version := fmt.Sprintf("%s-%d", pconfig.VcInfo.Pkgver, pconfig.VcInfo.Pkgrel)

		if exists, err := git.PackageUpdateBranchExists(pkgbase, version); err != nil {
			panic(err)
		} else if exists {
			slog.Info(fmt.Sprintf("Already have branch for updating pacakge %s to version %s. Skipping.", pkgbase, version))
			continue
		}

		slog.Info(fmt.Sprintf("Updating package %s", pkgbase))

		if cenv.IsCI() {
			if err := git.CreateAndSwitchToPackageUpdateBranch(pkgbase, version); err != nil {
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

			if err := git.Commit(fmt.Sprintf("Update %s at version %s", pkgbase, version)); err != nil {
				panic(err)
			}

			if err := git.PushPackageBranch(pkgbase, version); err != nil {
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
