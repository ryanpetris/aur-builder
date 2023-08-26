package cienv

import (
	"fmt"
	"github.com/go-git/go-git/v5"
)

type CiEnv interface {
	IsCI() bool
	CreatePR() error
	WriteBuildPackages(pkgbase []string) error
	SetGitCommitOptions(options *git.CommitOptions) error
	SetGitPushOptions(options *git.PushOptions) error
}

type DefaultCiEnv struct {
}

func (env DefaultCiEnv) IsCI() bool {
	return false
}

func (env DefaultCiEnv) CreatePR() error {
	return nil
}

func (env DefaultCiEnv) WriteBuildPackages(pkgbase []string) error {
	for _, pkgb := range pkgbase {
		fmt.Printf("%s needs update\n", pkgb)
	}

	return nil
}

func (env DefaultCiEnv) SetGitCommitOptions(options *git.CommitOptions) error {
	return nil
}

func (env DefaultCiEnv) SetGitPushOptions(options *git.PushOptions) error {
	return nil
}
