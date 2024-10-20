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
	"strings"
)

type ForgejoCiEnv struct {
}

func (env ForgejoCiEnv) IsCI() bool {
	if value := os.Getenv("GITEA_ACTIONS"); value == "true" {
		return true
	}

	return false
}

func (env ForgejoCiEnv) CreatePR() error {
	if !env.IsCI() {
		return errors.New("Not in CI environment")
	}

	branchBytes, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()

	if err != nil {
		return err
	}

	branchName := strings.SplitN(string(branchBytes), "\n", 2)[0]
	messageBytes, err := exec.Command("git", "log", "-n", "1", "--pretty=%B").Output()

	if err != nil {
		return err
	}

	title := strings.SplitN(string(messageBytes), "\n", 2)[0]

	data := map[string]any{
		"head":  branchName,
		"base":  "master",
		"title": title,
	}

	dataBytes, err := json.Marshal(data)

	if err != nil {
		return err
	}

	cmd := exec.Command(
		"curl", "-X", "POST",
		fmt.Sprintf("%s/repos/%s/pulls", os.Getenv("GITHUB_API_URL"), os.Getenv("GITHUB_REPOSITORY")),
		"--insecure",
		"--silent",
		"--user", fmt.Sprintf("me:%s", os.Getenv("GITHUB_TOKEN")),
		"--header", "Content-Type: application/json",
		"--data-raw", string(dataBytes),
	)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (env ForgejoCiEnv) WriteBuildPackages(pkgbase []string) error {
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

func (env ForgejoCiEnv) SetGitCommitOptions(options *git.CommitOptions) error {
	if ghActor := os.Getenv("GITHUB_ACTOR"); ghActor != "" {
		options.Author = &gitobject.Signature{
			Name:  ghActor,
			Email: fmt.Sprintf("%s@users.noreply.github.com", ghActor),
		}
	}

	return nil
}

func (env ForgejoCiEnv) SetGitPushOptions(options *git.PushOptions) error {
	if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
		options.Auth = &http.BasicAuth{Username: "me", Password: ghToken}
	}

	return nil
}
