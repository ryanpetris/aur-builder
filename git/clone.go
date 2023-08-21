package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/ryanpetris/aur-builder/config"
	"os"
	"path"
)

func CloneUpstream(pkgbase string, url string, branch string) error {
	upstreamPath := config.GetUpstreamPath(pkgbase)
	gitPath := path.Join(upstreamPath, ".git")

	if _, err := os.Stat(upstreamPath); err != nil {
		if err = os.RemoveAll(upstreamPath); err != nil {
			return err
		}
	}

	var branchRef plumbing.ReferenceName

	if branch != "" {
		branchRef = plumbing.NewBranchReferenceName(CleanBranchName(branch))
	}

	_, err := git.PlainClone(upstreamPath, false, &git.CloneOptions{
		URL:           url,
		Depth:         1,
		ReferenceName: branchRef,
		SingleBranch:  true,
	})

	if err != nil {
		return err
	}

	if err = os.RemoveAll(gitPath); err != nil {
		return err
	}

	return nil
}
