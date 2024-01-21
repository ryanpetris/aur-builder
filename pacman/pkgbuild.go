package pacman

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"os/exec"
	"strings"
)

//go:embed get_pkgbuild_var.sh
var getPkgbuildVarScript string

func GetPkgbuildVar(pkgbase string, varname string) (string, error) {
	mergedPath := config.GetMergedPath(pkgbase)

	var stdoutBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", getPkgbuildVarScript)
	cmd.Stdout = &stdoutBuf
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("PKGBUILD_PATH=%s/PKGBUILD", mergedPath))
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("PKGBUILD_VAR=%s", varname))

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return stdoutBuf.String(), nil
}

func GetPkgbuildVars(pkgbase string, varname string) ([]string, error) {
	out, err := GetPkgbuildVar(pkgbase, varname)

	if err != nil {
		return nil, err
	}

	var result []string

	for _, item := range strings.Split(out, "\n") {
		if item == "" {
			continue
		}

		result = append(result, item)
	}

	return result, nil
}
