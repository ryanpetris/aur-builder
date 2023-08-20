package impenv

import (
	"github.com/ryanpetris/aur-builder/arch"
	"github.com/ryanpetris/aur-builder/misc"
)

type ArchImportEnv struct {
}

func (ienv ArchImportEnv) GetPackageInfo(pkgname []string) ([]misc.PackageInfo, error) {
	data, err := arch.GetPackages(pkgname)

	if err != nil {
		return nil, err
	}

	var result []misc.PackageInfo

	for _, item := range data {
		result = append(result, misc.PackageInfo{
			Pkgbase:     item.Pkgbase,
			Pkgname:     item.Pkgname,
			FullVersion: item.GetFullVersion(),
		})
	}

	return result, nil
}

func (ienv ArchImportEnv) PackageExists(pkgbase string) (bool, error) {
	return arch.PackageExists(pkgbase)
}

func (ienv ArchImportEnv) PackageImport(pkgbase string) error {
	return arch.ClonePackage(pkgbase)
}
