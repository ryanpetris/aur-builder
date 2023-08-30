package arch

import (
	"errors"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"github.com/ryanpetris/aur-builder/git"
	"github.com/ryanpetris/aur-builder/pkg"
)

func ClonePackage(pkgbase string, version string) error {
	if exists, err := PackageExists(pkgbase); err != nil {
		return err
	} else if !exists {
		return errors.New(fmt.Sprintf("Package %s does not exist in the aur", pkgbase))
	}

	aurUrl := config.GetArchPackageGitUrl(pkgbase)

	if err := git.CloneUpstream(pkgbase, aurUrl, version); err != nil {
		return err
	}

	pconfig, err := pkg.LoadConfig(pkgbase)

	if err != nil {
		return err
	}

	pconfig.Source = "arch"

	if err := pconfig.CleanPkgrelBumpVersions(version); err != nil {
		return err
	}

	if err := pconfig.Write(pkgbase); err != nil {
		return err
	}

	return nil
}
