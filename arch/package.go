package arch

import (
	"bytes"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

type Package struct {
	Pkgbase string
	Pkgname string
	Epoch   int
	Pkgver  string
	Pkgrel  int
}

func (pkg *Package) Load(packagelines []string) error {
	var section string

	for _, line := range packagelines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "%") {
			section = line
			continue
		}

		switch section {
		case "%BASE%":
			pkg.Pkgbase = line
		case "%NAME%":
			pkg.Pkgname = line
		case "%VERSION%":
			version := line

			if parts := strings.SplitN(line, ":", 2); len(parts) > 1 {
				pkg.Epoch, _ = strconv.Atoi(parts[0])
				version = parts[1]
			}

			parts := strings.SplitN(version, "-", 2)

			pkg.Pkgver = parts[0]
			pkg.Pkgrel, _ = strconv.Atoi(parts[1])
		}
	}

	return nil
}

func (pkg *Package) GetFullVersion() string {
	if pkg.Epoch > 0 {
		return fmt.Sprintf("%d:%s-%d", pkg.Epoch, pkg.Pkgver, pkg.Pkgrel)
	}

	return fmt.Sprintf("%s-%d", pkg.Pkgver, pkg.Pkgrel)
}

func PackageExists(pkgbase string) (bool, error) {
	cmdText := `
(   tar -xvz -f /var/lib/pacman/sync/core.db  --wildcards '*/desc' --to-stdout 2>/dev/null \
 && tar -xvz -f /var/lib/pacman/sync/extra.db --wildcards '*/desc' --to-stdout 2>/dev/null) \
| grep -A 1 --no-group-separator -E '^%BASE%$' \
| grep -v '^%BASE%$' \
| sort \
| uniq
`
	var stdoutBuf = bytes.Buffer{}

	cmd := exec.Command("bash", "-c", cmdText)
	cmd.Stdout = &stdoutBuf

	if err := cmd.Run(); err != nil {
		return false, err
	}

	packages := strings.Split(stdoutBuf.String(), "\n")

	for _, pkg := range packages {
		if pkg == pkgbase {
			return true, nil
		}
	}

	return false, nil
}

func GetPackages(pkgname []string) ([]Package, error) {
	cmdText := `
   tar -xvz -f /var/lib/pacman/sync/core.db  --wildcards '*/desc' --to-stdout 2>/dev/null \
&& tar -xvz -f /var/lib/pacman/sync/extra.db --wildcards '*/desc' --to-stdout 2>/dev/null
`
	var stdoutBuf = bytes.Buffer{}

	cmd := exec.Command("bash", "-c", cmdText)
	cmd.Stdout = &stdoutBuf

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(stdoutBuf.String(), "\n")

	nextLineIsPkgname := false
	skipCurrent := true
	var current []string
	var packages []Package

	processCurrent := func() error {
		if len(current) > 0 && !skipCurrent {
			newPackage := Package{}

			if err := newPackage.Load(current); err != nil {
				return err
			}

			packages = append(packages, newPackage)
		}

		current = nil
		nextLineIsPkgname = false
		skipCurrent = true

		return nil
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		if line == "%FILENAME%" {
			if err := processCurrent(); err != nil {
				return nil, err
			}
		}

		if nextLineIsPkgname {
			nextLineIsPkgname = false

			if slices.Contains(pkgname, line) {
				skipCurrent = false
			}
		}

		if line == "%NAME%" {
			nextLineIsPkgname = true
		}

		current = append(current, line)
	}

	if err := processCurrent(); err != nil {
		return nil, err
	}

	return packages, nil
}

func GetPackageVersion(pkgbase string) (string, error) {
	cmdText := `
   tar -xvz -f /var/lib/pacman/sync/core.db --wildcards '*/desc' --to-stdout 2>/dev/null \
&& tar -xvz -f /var/lib/pacman/sync/extra.db --wildcards '*/desc' --to-stdout 2>/dev/null
`
	var stdoutBuf = bytes.Buffer{}

	cmd := exec.Command("bash", "-c", cmdText)
	cmd.Stdout = &stdoutBuf

	if err := cmd.Run(); err != nil {
		return "", err
	}

	lines := strings.Split(stdoutBuf.String(), "\n")
	var section string
	var foundPackage bool

	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "%") {
			section = line
		}

		if section == "%BASE%" && line == pkgbase {
			foundPackage = true
		} else if section == "%VERSION%" && foundPackage {
			return line, nil
		}
	}

	return "", nil
}
