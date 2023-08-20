package pkg

import (
	"errors"
	"github.com/ryanpetris/aur-builder/config"
	"os"
)

func PackageExists(pkgbase string) (bool, error) {
	basePath := config.GetPackagePath(pkgbase)

	if _, err := os.Stat(basePath); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func GetPackages() ([]string, error) {
	basePath := config.GetBasePath()
	entries, err := os.ReadDir(basePath)

	if err != nil {
		return nil, err
	}

	var packages []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		packages = append(packages, entry.Name())
	}

	return packages, nil
}
