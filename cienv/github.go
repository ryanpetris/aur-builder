package cienv

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

	data := fmt.Sprintf("packages=%s", dataJson)

	if ghOutputFile := os.Getenv("GITHUB_OUTPUT"); ghOutputFile != "" {
		ghOutput, err := os.OpenFile(ghOutputFile, os.O_APPEND|os.O_WRONLY, 0666)

		if err != nil {
			return err
		}

		defer ghOutput.Close()

		if _, err = ghOutput.WriteString(data); err != nil {
			return err
		}
	} else {
		fmt.Println(data)
	}

	return nil
}

func (env GithubCiEnv) SetGitCommitOptions(options *git.CommitOptions) error {
	if ghActor := os.Getenv("GITHUB_ACTOR"); ghActor != "" {
		options.Author = &gitobject.Signature{
			Name:  ghActor,
			Email: fmt.Sprintf("%s@users.noreply.github.com", ghActor),
		}
	}

	return nil
}

func (env GithubCiEnv) SetGitPushOptions(options *git.PushOptions) error {
	if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
		options.Auth = &http.BasicAuth{Username: "me", Password: ghToken}
	}

	return nil
}
