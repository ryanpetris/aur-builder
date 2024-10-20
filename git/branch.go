package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/ryanpetris/aur-builder/cienv"
)
import "github.com/go-git/go-git/v5/plumbing"

func PackageUpdateBranchExists(pkgbase string, pkgver string) (bool, error) {
	remotePath := fmt.Sprintf("origin/packages/%s/%s", pkgbase, CleanTagName(pkgver))
	repo, err := git.PlainOpen(".")

	if err != nil {
		return false, err
	}

	_, err = repo.ResolveRevision(plumbing.Revision(remotePath))

	return err == nil, nil
}

func CreateAndSwitchToPackageUpdateBranch(pkgbase string, pkgver string) error {
	branchRef := plumbing.NewBranchReferenceName(fmt.Sprintf("packages/%s/%s", pkgbase, CleanTagName(pkgver)))
	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	headRef, err := repo.Head()

	if err != nil {
		return err
	}

	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))

	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Keep:   true,
	}); err != nil {
		return err
	}

	return nil
}

func SwitchToMaster() error {
	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("master"),
	}); err != nil {
		return err
	}

	return nil
}

func PushPackageBranch(pkgbase string, pkgver string) error {
	branchRef := fmt.Sprintf("packages/%s/%s", pkgbase, CleanTagName(pkgver))
	repo, err := git.PlainOpen(".")

	if err != nil {
		return err
	}

	pushOptions := git.PushOptions{
		RemoteName:      "origin",
		RefSpecs:        []config.RefSpec{config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/heads/%s", branchRef, branchRef))},
		InsecureSkipTLS: insecureSkipTls,
	}

	ce := cienv.FindCiEnv()

	if err := ce.SetGitPushOptions(&pushOptions); err != nil {
		return err
	}

	if err := repo.Push(&pushOptions); err != nil {
		return err
	}

	return nil
}
