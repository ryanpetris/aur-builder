package pacman

import (
	"bytes"
	"fmt"
	"github.com/Jguer/go-alpm/v2"
	"log/slog"
	"os/exec"
	"strings"
)

func GetPackageVersion(pkgname string) (string, error) {
	slog.Debug(fmt.Sprintf("Looking up version for package %s", pkgname))

	var stdoutBuf bytes.Buffer

	cmd := exec.Command("pacman", "-Si", pkgname)
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()

	if err != nil {
		slog.Debug(fmt.Sprintf("Error running pacman command %s", err))
		return "", nil
	}

	for _, line := range strings.Split(stdoutBuf.String(), "\n") {
		if strings.HasPrefix(line, "Version") {
			parts := strings.SplitN(line, ":", 2)

			return strings.Trim(parts[1], " "), nil
		}
	}

	slog.Debug(fmt.Sprintf("Version not found in output for package %s", pkgname))

	return "", nil
}

func IsVersionNewer(oldVersion string, newVersion string) (bool, error) {
	result := alpm.VerCmp(oldVersion, newVersion)

	return result < 0, nil
}
