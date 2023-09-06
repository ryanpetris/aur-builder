package pkg

import (
	"bytes"
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"github.com/ryanpetris/aur-builder/misc"
	"github.com/ryanpetris/aur-builder/pacman"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
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

	if pconfig.Overrides.ClearReplaces {
		err := processClearReplaces(pkgbase)

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

func processClearReplaces(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear replaces override for pkgbase %s", pkgbase))

	appendText := `unset replaces`

	return appendPkgbuild(pkgbase, appendText)
}

func processClearSignatures(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear signatures override for pkgbase %s", pkgbase))

	type arrayDef struct {
		Name     string
		OrigName string
		Arch     string
		Items    map[string]string
	}

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	var stdoutBuf = bytes.Buffer{}

	cmd := exec.Command("bash", "-c", fmt.Sprintf(`source "%s"; declare -p | grep -E '^declare -a (source|[a-z0-9]+sums)(_[a-z0-9_]+)?='`, pkgbuildPath))
	cmd.Stdout = &stdoutBuf

	if err := cmd.Run(); err != nil {
		return err
	}

	re, err := regexp.Compile(`\[(?P<index>[0-9]+)]="(?P<value>(?:[^\\"]|\\"|\\[^"])*?)"`)

	if err != nil {
		return err
	}

	var parsedArrays []arrayDef

	for _, line := range strings.Split(stdoutBuf.String(), "\n") {
		if line == "" {
			continue
		}

		declareParts := strings.SplitN(line, " ", 3)
		varParts := strings.SplitN(declareParts[2], "=", 2)
		nameParts := strings.SplitN(varParts[0], "_", 2)

		def := arrayDef{
			Name:     nameParts[0],
			OrigName: varParts[0],
			Items:    map[string]string{},
		}

		if len(nameParts) > 1 {
			def.Arch = nameParts[1]
		}

		matches := re.FindAllStringSubmatch(varParts[1], -1)

		for _, match := range matches {
			matchMap := misc.RegexMapMatchByGroup(re, match)

			def.Items[matchMap["index"]] = matchMap["value"]
		}

		parsedArrays = append(parsedArrays, def)
	}

	appendLines := []string{"unset validpgpkeys"}
	var affectedArrays []string

	for _, item := range parsedArrays {
		if item.Name != "source" {
			continue
		}

		for index, value := range item.Items {
			if strings.HasSuffix(value, ".sig") {
				if !slices.Contains(affectedArrays, item.OrigName) {
					affectedArrays = append(affectedArrays, item.OrigName)
				}

				appendLines = append(appendLines, fmt.Sprintf("unset %s[%s]", item.OrigName, index))

				for _, sumItem := range parsedArrays {
					if !strings.HasSuffix(sumItem.Name, "sums") {
						continue
					}

					if sumItem.Arch != item.Arch {
						continue
					}

					if !slices.Contains(affectedArrays, sumItem.OrigName) {
						affectedArrays = append(affectedArrays, sumItem.OrigName)
					}

					appendLines = append(appendLines, fmt.Sprintf("unset %s[%s]", sumItem.OrigName, index))
				}
			}
		}
	}

	for _, item := range affectedArrays {
		appendLines = append(appendLines, fmt.Sprintf(`mapfile -t %s < <(IFS=$'\n'; echo "${%s[*]}")`, item, item))
	}

	appendText := strings.Join(appendLines, "\n")

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

	anyvarRegex, err := regexp.Compile(`^[a-zA-Z0-9@._+-]+\s*=`)

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
