package cli

import (
	"flag"
	"fmt"
	"github.com/ryanpetris/aur-builder/cienv"
	"github.com/ryanpetris/aur-builder/git"
	"github.com/ryanpetris/aur-builder/impenv"
	"github.com/ryanpetris/aur-builder/pacman"
	"github.com/ryanpetris/aur-builder/pkg"
	"strings"
)

func ImportMain(args []string) {
	cmd := flag.NewFlagSet("import", flag.ExitOnError)

	cmdSource := cmd.String("source", "", "package source (aur, arch)")
	cmdPackage := cmd.String("package", "", "name of package to import")

	if err := cmd.Parse(args[1:]); err != nil {
		panic(err)
	}

	if *cmdSource == "" {
		panic("--source is required")
	}

	if *cmdPackage == "" {
		panic("--package is required")
	}

	var ienv impenv.ImportEnv

	switch strings.ToLower(*cmdSource) {
	case "aur":
		ienv = impenv.AurImportEnv{}
	case "arch":
		ienv = impenv.ArchImportEnv{}
	default:
		panic(fmt.Sprintf("Invalid source: %s", *cmdSource))
	}

	pkgbase := strings.ToLower(*cmdPackage)

	if exists, err := pkg.PackageExists(pkgbase); err != nil {
		panic(err)
	} else if exists {
		panic(fmt.Sprintf("Package %s already imported", pkgbase))
	}

	cenv := cienv.FindCiEnv()

	if cenv.IsCI() {
		if err := git.CreateAndSwitchToPackageUpdateBranch(pkgbase, "0"); err != nil {
			panic(err)
		}
	}

	if exists, err := ienv.PackageExists(pkgbase); err != nil {
		panic(err)
	} else if !exists {
		panic(fmt.Sprintf("Package %s does not exist in source %s", pkgbase, *cmdSource))
	}

	if err := ienv.PackageImport(pkgbase, ""); err != nil {
		panic(err)
	}

	pconfig, err := pkg.LoadConfig(pkgbase)

	if err != nil {
		panic(err)
	}

	if updated, err := pconfig.GenVcsInfo(pkgbase); err != nil {
		panic(err)
	} else if updated {
		if err := pconfig.Write(pkgbase); err != nil {
			panic(err)
		}
	}

	if err := pconfig.ClearMerge(pkgbase); err != nil {
		panic(err)
	}

	if err := pconfig.Merge(pkgbase, false); err != nil {
		panic(err)
	}

	if cenv.IsCI() {
		pkginfo, err := pacman.LoadPkgInfo(pkgbase)
		pkgver := pkginfo.GetFullVersion()

		if err != nil {
			panic(err)
		}

		if err := git.CreateAndSwitchToPackageUpdateBranch(pkgbase, pkgver); err != nil {
			panic(err)
		}

		if err := git.AddAll(); err != nil {
			panic(err)
		}

		if err := git.Commit(fmt.Sprintf("Add %s at version %s", pkgbase, pkgver)); err != nil {
			panic(err)
		}

		if err := git.PushPackageBranch(pkgbase, pkgver); err != nil {
			panic(err)
		}

		if err := cenv.CreatePR(); err != nil {
			panic(err)
		}

		if err := git.SwitchToMaster(); err != nil {
			panic(err)
		}
	} else {
		if err := pconfig.ClearMerge(pkgbase); err != nil {
			panic(err)
		}
	}
}
