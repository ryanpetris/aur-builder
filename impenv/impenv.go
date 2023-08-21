package impenv

import "github.com/ryanpetris/aur-builder/misc"

type ImportEnv interface {
	GetPackageInfo(pkgname []string) ([]misc.PackageInfo, error)
	PackageExists(pkgbase string) (bool, error)
	PackageImport(pkgbase string, version string) error
}
