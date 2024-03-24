package cli

import (
	"flag"
	"fmt"
	"github.com/ryanpetris/aur-builder/cienv"
	"github.com/ryanpetris/aur-builder/git"
	"github.com/ryanpetris/aur-builder/pkg"
	"log/slog"
	"os"
)

func UpdateVcsMain(args []string) {
	cmd := flag.NewFlagSet("update-vcs", flag.ExitOnError)

	cmdPackage := cmd.String("package", "", "name of package to update")
	cmdAll := cmd.Bool("all", false, "check all packages")

	if err := cmd.Parse(args[1:]); err != nil {
		panic(err)
	}

	if *cmdPackage != "" && *cmdAll {
		slog.Error("--package and --all options are mutually-exclusive.")
		os.Exit(1)
	}

	cenv := cienv.FindCiEnv()
	allPackages, err := pkg.GetPackages()

	if err != nil {
		panic(err)
	}

	for _, pkgbase := range allPackages {
		if *cmdPackage != "" && *cmdPackage != pkgbase {
			continue
		}

		pconfig, err := pkg.LoadConfig(pkgbase)

		if err != nil {
			panic(err)
		}

		if *cmdPackage == "" {
			if pconfig.Ignore {
				continue
			}

			if pconfig.Vcs == nil && !*cmdAll {
				continue
			}
		} else if *cmdPackage != pkgbase {
			continue
		}

		slog.Info(fmt.Sprintf("Checking package %s for VCS updates...", pkgbase))

		updated, err := pconfig.GenVcsInfo(pkgbase)

		if err != nil {
			panic(err)
		}

		if !updated {
			continue
		}

		version := fmt.Sprintf("%s-%d", pconfig.Vcs.Pkgver, pconfig.Vcs.Pkgrel)

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

		if err := pconfig.ClearMerge(pkgbase); err != nil {
			panic(err)
		}

		if cenv.IsCI() {
			if err := pconfig.Merge(pkgbase, true); err != nil {
				panic(err)
			}

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
