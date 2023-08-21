package impenv

import (
	"github.com/ryanpetris/aur-builder/aur"
	"github.com/ryanpetris/aur-builder/misc"
)

type AurImportEnv struct {
}

func (ienv AurImportEnv) GetPackageInfo(pkgname []string) ([]misc.PackageInfo, error) {
	data, err := aur.GetPackageInfos(pkgname)

	if err != nil {
		return nil, err
	}

	var result []misc.PackageInfo

	for _, item := range data {
		result = append(result, misc.PackageInfo{
			Pkgbase:     item.PackageBase,
			Pkgname:     item.Name,
			FullVersion: item.Version,
		})
	}

	return result, nil
}

func (ienv AurImportEnv) PackageExists(pkgbase string) (bool, error) {
	return aur.PackageExists(pkgbase)
}

func (ienv AurImportEnv) PackageImport(pkgbase string, version string) error {
	return aur.ClonePackage(pkgbase)
}
