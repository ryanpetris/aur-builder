package cienv

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
)

type CiEnv interface {
	IsCI() bool
	CreatePR() error
	WriteBuildPackages(pkgbase []string) error
	SetCommitOptions(options *git.CommitOptions) error
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

func (env DefaultCiEnv) SetCommitOptions(options *git.CommitOptions) error {
	return errors.New("Not in CI environment")
}
