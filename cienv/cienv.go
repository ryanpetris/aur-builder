package cienv

import (
	"errors"
	"fmt"
)

type CiEnv interface {
	IsCI() bool
	CreatePR() error
	WriteBuildPackages(pkgbase []string) error
}

type DefaultCiEnv struct {
}

func (env DefaultCiEnv) IsCI() bool {
	return false
}

func (env DefaultCiEnv) CreatePR() error {
	return errors.New("Not in CI environment")
}

func (env DefaultCiEnv) WriteBuildPackages(pkgbase []string) error {
	for _, pkgb := range pkgbase {
		fmt.Printf("%s needs update\n", pkgb)
	}

	return nil
}
