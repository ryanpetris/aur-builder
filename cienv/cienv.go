package cienv

import "errors"

type CiEnv interface {
	IsCI() bool
	CreatePR() error
}

type DefaultCiEnv struct {
}

func (env DefaultCiEnv) IsCI() bool {
	return false
}

func (env DefaultCiEnv) CreatePR() error {
	return errors.New("Not in CI environment")
}
