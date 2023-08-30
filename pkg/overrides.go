package pkg

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"github.com/ryanpetris/aur-builder/misc"
	"github.com/ryanpetris/aur-builder/pacman"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

func (pconfig *PackageConfig) ProcessOverrides(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing overrides for pkgbase %s", pkgbase))

	// First run functions that manipulate the PKGBUILD

	if pconfig.Overrides.RenamePackage != nil {
		err := processRenamePackage(pkgbase, pconfig.Overrides.RenamePackage)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.ReplacePkgbuild != nil {
		err := processReplacePkgbuild(pkgbase, pconfig.Overrides.ReplacePkgbuild)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.RenameFunction != nil {
		err := processRenameFunction(pkgbase, pconfig.Overrides.RenameFunction)

		if err != nil {
			return err
		}
	}

	// Then run functions that merely append to the PKGBUILD

	if pconfig.Overrides.BumpPkgrel != nil {
		err := processBumpPkgrel(pconfig, pkgbase)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.ClearConflicts {
		err := processClearConflicts(pkgbase)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.ClearDependsVersions {
		err := processClearDependsVersions(pkgbase)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.ClearPkgverFunc {
		err := processClearPkgverFunc(pkgbase)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.ClearProvides {
		err := processClearProvides(pkgbase)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.ClearSignatures {
		err := processClearSignatures(pkgbase)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.AppendPkgbuild != "" {
		err := processAppendPkgbuild(pkgbase, pconfig.Overrides.AppendPkgbuild)

		if err != nil {
			return err
		}
	}

	// Then run functions that don't touch the PKGBUILD at all

	if pconfig.Overrides.DeleteFile != nil {
		err := processDeleteFile(pkgbase, pconfig.Overrides.DeleteFile)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.RenameFile != nil {
		err := processRenameFile(pkgbase, pconfig.Overrides.RenameFile)

		if err != nil {
			return err
		}
	}

	return nil
}

func processAppendPkgbuild(pkgbase string, appendText string) error {
	slog.Debug(fmt.Sprintf("Processing append pkgbuild override for pkgbase %s", pkgbase))

	return appendPkgbuild(pkgbase, appendText)
}

func processBumpPkgrel(pconfig *PackageConfig, pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing pkgrel bump overrides for pkgbase %s", pkgbase))

	tmpl := template.New("t")
	tmpl, err := tmpl.Parse(`
if [ "$pkgver" = "{{ .Version }}" ]; then
    pkgrel=$((pkgrel + {{ .Bump }}))
fi
`)

	if err != nil {
		return err
	}

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuild, err := os.OpenFile(pkgbuildPath, os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		return err
	}

	defer pkgbuild.Close()

	for version, bump := range pconfig.Overrides.BumpPkgrel {
		slog.Debug(fmt.Sprintf("Adding %d to pkgrel for pkgbase %s version %s", bump, pkgbase, version))

		err = tmpl.Execute(pkgbuild, map[string]string{
			"Version": version,
			"Bump":    strconv.Itoa(bump),
		})

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func processClearConflicts(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear conflicts override for pkgbase %s", pkgbase))

	appendText := `unset conflicts`

	return appendPkgbuild(pkgbase, appendText)
}

func processClearDependsVersions(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear depends versions override for pkgbase %s", pkgbase))

	appendText := `mapfile -t depends < <((IFS=$'\n'; echo "${depends[*]}") | sed -E 's/[<>=].*$//' | sort | uniq)`

	return appendPkgbuild(pkgbase, appendText)
}

func processClearPkgverFunc(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear pkgver function override for pkgbase %s", pkgbase))

	appendText := `unset -f pkgver`

	return appendPkgbuild(pkgbase, appendText)
}

func processClearProvides(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear provides override for pkgbase %s", pkgbase))

	appendText := `unset provides`

	return appendPkgbuild(pkgbase, appendText)
}

func processClearSignatures(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear signatures override for pkgbase %s", pkgbase))

	appendText := `
unset validpgpkeys

_new_source=()
_new_b2sums=()
_new_sha512sums=()
_new_sha384sums=()
_new_sha256sums=()
_new_sha224sums=()
_new_sha1sums=()
_new_md5sums=()
_new_cksums=()

for i in "${!source[@]}"; do
    if [[ "${source[$i]}" == *.sig ]]; then
        continue
    fi

    _new_source+=("${source[$i]}")

	if [[ "${#b2sums[@]}" != "0" ]]; then
    	_new_b2sums+=("${b2sums[$i]}")
	fi

	if [[ "${#sha512sums[@]}" != "0" ]]; then
    	_new_sha512sums+=("${sha512sums[$i]}")
	fi

	if [[ "${#sha384sums[@]}" != "0" ]]; then
    	_new_sha384sums+=("${sha384sums[$i]}")
	fi

	if [[ "${#sha256sums[@]}" != "0" ]]; then
    	_new_sha256sums+=("${sha256sums[$i]}")
	fi

	if [[ "${#sha224sums[@]}" != "0" ]]; then
    	_new_sha224sums+=("${sha224sums[$i]}")
	fi

	if [[ "${#sha1sums[@]}" != "0" ]]; then
    	_new_sha1sums+=("${sha1sums[$i]}")
	fi

	if [[ "${#md5sums[@]}" != "0" ]]; then
    	_new_md5sums+=("${md5sums[$i]}")
	fi

	if [[ "${#cksums[@]}" != "0" ]]; then
    	_new_b2sums+=("${cksums[$i]}")
	fi
done

source=("${_new_source[@]}")

if [[ "${#b2sums[@]}" != "0" ]]; then
	b2sums=("${_new_b2sums[@]}")
fi

if [[ "${#sha512sums[@]}" != "0" ]]; then
	sha512sums=("${_new_sha512sums[@]}")
fi

if [[ "${#sha384sums[@]}" != "0" ]]; then
	sha384sums=("${_new_sha384sums[@]}")
fi

if [[ "${#sha256sums[@]}" != "0" ]]; then
	sha256sums=("${_new_sha256sums[@]}")
fi

if [[ "${#sha224sums[@]}" != "0" ]]; then
	sha224sums=("${_new_sha224sums[@]}")
fi

if [[ "${#sha1sums[@]}" != "0" ]]; then
	sha1sums=("${_new_sha1sums[@]}")
fi

if [[ "${#md5sums[@]}" != "0" ]]; then
	md5sums=("${_new_md5sums[@]}")
fi

if [[ "${#cksums[@]}" != "0" ]]; then
	cksums=("${_new_cksums[@]}")
fi

unset _new_source
unset _new_b2sums
unset _new_sha512sums
unset _new_sha384sums
unset _new_sha256sums
unset _new_sha224sums
unset _new_sha1sums
unset _new_md5sums
unset _new_cksums
`

	return appendPkgbuild(pkgbase, appendText)
}

func processDeleteFile(pkgbase string, files []string) error {
	slog.Debug(fmt.Sprintf("Processing delete file override for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)

	for _, item := range files {
		filePath := path.Join(mergedPath, item)

		if err := os.RemoveAll(filePath); err != nil {
			return err
		}
	}

	return nil
}

func processRenameFile(pkgbase string, overrides []PackageConfigOverrideFromTo) error {
	slog.Debug(fmt.Sprintf("Processing move file override for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)

	for _, item := range overrides {
		fromPath := path.Join(mergedPath, item.From)
		toPath := path.Join(mergedPath, item.To)

		if err := os.Rename(fromPath, toPath); err != nil {
			return err
		}
	}

	return nil
}

func processRenameFunction(pkgbase string, overrides []PackageConfigOverrideFromTo) error {
	slog.Debug(fmt.Sprintf("Processing rename function override for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuildBytes, err := os.ReadFile(pkgbuildPath)

	if err != nil {
		return err
	}

	pkgbuild := string(pkgbuildBytes)

	namechangemap := map[string]string{}

	for _, item := range overrides {
		namechangemap[item.From] = item.To
	}

	if result, err := replaceFunctionNames(pkgbuild, namechangemap); err != nil {
		return err
	} else {
		pkgbuild = result
	}

	if err := os.WriteFile(pkgbuildPath, []byte(pkgbuild), 0666); err != nil {
		return err
	}

	return nil
}

func processRenamePackage(pkgbase string, overrides []PackageConfigOverrideFromTo) error {
	slog.Debug(fmt.Sprintf("Processing rename package override for pkgbase %s", pkgbase))

	if err := pacman.GenSrcInfo(pkgbase); err != nil {
		return err
	}

	srcinfos, err := pacman.LoadSrcinfo(pkgbase)

	if err != nil {
		return err
	}

	var pkgnames []string
	namechangemap := map[string]string{}
	functypenames := []string{"package", "prepare", "build", "check"}

	for _, srcinfo := range srcinfos {
		found := false

		for _, override := range overrides {
			if override.From == srcinfo.Pkgname || (override.From == "" && srcinfo.Pkgname == pkgbase) {
				if override.To != "" {
					pkgnames = append(pkgnames, override.To)

					for _, functypename := range functypenames {
						namechangemap[fmt.Sprintf("%s_%s", functypename, srcinfo.Pkgname)] = fmt.Sprintf("%s_%s", functypename, override.To)
					}
				}

				found = true
				break
			}
		}

		if !found {
			pkgnames = append(pkgnames, srcinfo.Pkgname)
		}
	}

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuildBytes, err := os.ReadFile(pkgbuildPath)

	if err != nil {
		return err
	}

	pkgbuild := string(pkgbuildBytes)

	if result, err := replacePkgname(pkgbuild, pkgnames); err != nil {
		return err
	} else {
		pkgbuild = result
	}

	if result, err := replaceFunctionNames(pkgbuild, namechangemap); err != nil {
		return err
	} else {
		pkgbuild = result
	}

	if err := os.WriteFile(pkgbuildPath, []byte(pkgbuild), 0666); err != nil {
		return err
	}

	return nil
}

func processReplacePkgbuild(pkgbase string, overrides []PackageConfigOverrideFromTo) error {
	slog.Debug(fmt.Sprintf("Processing replace pkgbuild override for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuildBytes, err := os.ReadFile(pkgbuildPath)

	if err != nil {
		return err
	}

	pkgbuild := string(pkgbuildBytes)

	for _, item := range overrides {
		re, err := regexp.Compile(item.From)

		if err != nil {
			return err
		}

		pkgbuild = re.ReplaceAllString(pkgbuild, item.To)
	}

	if err := os.WriteFile(pkgbuildPath, []byte(pkgbuild), 0666); err != nil {
		return err
	}

	return nil
}

func appendPkgbuild(pkgbase string, appendText string) error {
	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuild, err := os.OpenFile(pkgbuildPath, os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		return err
	}

	defer pkgbuild.Close()

	if _, err = pkgbuild.WriteString(fmt.Sprintf("\n%s\n", appendText)); err != nil {
		return err
	}

	return nil
}

func replacePkgname(pkgbuild string, pkgnames []string) (string, error) {
	pkgnameRegex, err := regexp.Compile(`^pkgname\s*=`)

	if err != nil {
		return "", err
	}

	anyvarRegex, err := regexp.Compile(`^[a-zA-Z0-9]+\s*=`)

	if err != nil {
		return "", err
	}

	oldlines := strings.Split(pkgbuild, "\n")
	var newlines []string
	foundPkgname := false
	foundNext := false

	for _, line := range oldlines {
		if !foundPkgname && pkgnameRegex.MatchString(line) {
			newnames := strings.Join(pkgnames, " ")

			if len(pkgnames) > 1 {
				newnames = fmt.Sprintf("(%s)", newnames)
			}

			newnames = fmt.Sprintf("pkgname=%s", newnames)

			newlines = append(newlines, newnames)
			foundPkgname = true
			continue
		}

		if foundPkgname && !foundNext {
			if anyvarRegex.MatchString(line) {
				foundNext = true
			} else {
				continue
			}
		}

		newlines = append(newlines, line)
	}

	return strings.Join(newlines, "\n"), nil
}

func replaceFunctionNames(pkgbuild string, namechangemap map[string]string) (string, error) {
	re, err := regexp.Compile(`(?m)^(?P<name>[a-zA-Z0-9@._+-]+)(?P<end>\s*\()`)

	if err != nil {
		return "", nil
	}

	return re.ReplaceAllStringFunc(pkgbuild, func(match string) string {
		parts := misc.RegexGetMatchByGroup(re, match)

		if newname, hasKey := namechangemap[parts["name"]]; hasKey {
			return fmt.Sprintf("%s%s", newname, parts["end"])
		}

		return match
	}), nil
}
