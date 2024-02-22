package pacman

import (
	"bytes"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"log/slog"
	"os"
	"os/exec"
	"path"
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

func GenSrcInfo(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Generating .SRCINFO for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	srcinfoPath := path.Join(mergedPath, ".SRCINFO")
	var stdoutBuf bytes.Buffer

	cmd := exec.Command("makepkg", "--printsrcinfo")
	cmd.Dir = mergedPath
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()

	if err != nil {
		slog.Error(fmt.Sprintf("Failed generating .SRCINFO for pkgbase %s", pkgbase))

		return err
	}

	err = os.WriteFile(srcinfoPath, stdoutBuf.Bytes(), 0666)

	if err != nil {
		slog.Error(fmt.Sprintf("Failed writing .SRCINFO for pkgbase %s", pkgbase))

		return err
	}

	return nil
}
