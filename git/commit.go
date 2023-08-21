package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
	"os/exec"
)

func AddAll() error {
	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return err
	}

	if _, err = worktree.Add("."); err != nil {
		return err
	}

	return nil
}

func Commit(message string) error {
	var cmdParts []string

	if ghActor := os.Getenv("GITHUB_ACTOR"); ghActor != "" {
		cmdParts = append(cmdParts, "-c", fmt.Sprintf("user.name=%s", ghActor), "-c", fmt.Sprintf("user.email=%s@users.noreply.github.com", ghActor))
	}

	cmdParts = append(cmdParts, "commit", "-m", message)

	cmd := exec.Command("git", cmdParts[:]...)

	return cmd.Run()
}
