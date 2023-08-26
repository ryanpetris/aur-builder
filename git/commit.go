package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/ryanpetris/aur-builder/cienv"
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
	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return err
	}

	options := git.CommitOptions{}
	ce := cienv.FindCiEnv()

	if err := ce.SetGitCommitOptions(&options); err != nil {
		return err
	}

	if _, err = worktree.Commit(message, &options); err != nil {
		return err
	}

	return nil
}
