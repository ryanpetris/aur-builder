package cienv

import (
	"errors"
	"os"
	"os/exec"
)

type GithubCiEnv struct {
}

func (env GithubCiEnv) IsCI() bool {
	if value := os.Getenv("GITHUB_RUN_ID"); value != "" {
		return true
	}

	return false
}

func (env GithubCiEnv) CreatePR() error {
	if !env.IsCI() {
		return errors.New("Not in CI environment")
	}

	cmd := exec.Command("gh", "pr", "create", "--fill", "--base", "master")

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
