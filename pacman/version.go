package pacman

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func IsVersionNewer(oldVersion string, newVersion string) (bool, error) {
	outBuf := &bytes.Buffer{}

	cmd := exec.Command("vercmp", oldVersion, newVersion)
	cmd.Stdout = outBuf

	err := cmd.Run()

	if err != nil {
		fmt.Println(outBuf.String())
		return false, err
	}

	valueStr := strings.SplitN(outBuf.String(), "\n", 2)[0]
	value, err := strconv.Atoi(valueStr)

	if err != nil {
		return false, err
	}

	return value < 0, nil
}
