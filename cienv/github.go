package cienv

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
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

func (env GithubCiEnv) WriteBuildPackages(pkgbase []string) error {
	dataJson, err := json.Marshal(pkgbase)

	if err != nil {
		return err
	}

	fmt.Printf("packages=%s", dataJson)

	return nil
}

func (env GithubCiEnv) SetCommitOptions(options *git.CommitOptions) error {
	if ghActor := os.Getenv("GITHUB_ACTOR"); ghActor != "" {
		options.Author = &gitobject.Signature{
			Name:  ghActor,
			Email: fmt.Sprintf("%s@users.noreply.github.com", ghActor),
		}
	}

	return nil
}
