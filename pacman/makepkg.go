package pacman

import (
	"bytes"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"log/slog"
	"os/exec"
)

func DownloadSources(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Downloading sources for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	outBuf := &bytes.Buffer{}

	cmd := exec.Command("makepkg", "--noprepare", "--nobuild", "--nodeps", "--holdver")
	cmd.Dir = mergedPath
	cmd.Stdout = outBuf
	cmd.Stderr = outBuf

	err := cmd.Run()

	if err != nil {
		slog.Error(fmt.Sprintf("Failed downloading sources for pkgbase %s\n%s", pkgbase, outBuf.String()))

		return err
	}

	return nil
}
