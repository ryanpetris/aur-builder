package cli

import (
	"flag"
	"github.com/ryanpetris/aur-builder/pkg"
)

func FormatConfigMain(args []string) {
	cmd := flag.NewFlagSet("formatconfig", flag.ExitOnError)

	cmdPackage := cmd.String("formatconfig", "", "name of package to format the config.yaml file for")

	if err := cmd.Parse(args[1:]); err != nil {
		panic(err)
	}

	var packages []string

	if *cmdPackage != "" {
		packages = []string{*cmdPackage}
	} else {
		pkgs, err := pkg.GetPackages()

		if err != nil {
			panic(err)
		}

		packages = pkgs
	}

	for _, pkgbase := range packages {
		if exists, err := pkg.ConfigExists(pkgbase); err != nil {
			panic(err)
		} else if !exists {
			continue
		}

		if config, err := pkg.LoadConfig(pkgbase); err != nil {
			panic(err)
		} else if err := config.Write(pkgbase); err != nil {
			panic(err)
		}
	}
}
