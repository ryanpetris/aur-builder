package cli

import (
	"flag"
	"github.com/ryanpetris/aur-builder/pacman"
	"github.com/ryanpetris/aur-builder/pkg"
	"sync"
)

func PrepareMain(args []string) {
	cmd := flag.NewFlagSet("prepare", flag.ExitOnError)

	cmdPackage := cmd.String("package", "", "name of package to prepare")
	cmdNoVcs := cmd.Bool("no-vcs", false, "don't process vcs overrides")

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

	var wg sync.WaitGroup

	for _, pkgbase := range packages {
		wg.Add(1)
		go processPackage(pkgbase, &wg, !*cmdNoVcs)
	}

	wg.Wait()
}

func processPackage(pkgbase string, wg *sync.WaitGroup, processVcs bool) {
	defer wg.Done()
	pconfig, err := pkg.LoadConfig(pkgbase)

	if err != nil {
		panic(err)
	}

	if err := pconfig.Merge(pkgbase, processVcs); err != nil {
		panic(err)
	}

	if err := pacman.GenSrcInfo(pkgbase); err != nil {
		panic(err)
	}
}
