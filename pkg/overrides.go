package pkg

import (
	"bytes"
	"errors"
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

	if pconfig.Overrides == nil {
		return nil
	}

	// First run functions that manipulate the PKGBUILD

	if pconfig.Overrides.RenamePackage != nil {
		err := processRenamePackage(pkgbase, pconfig.Overrides.RenamePackage)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.ModifySection != nil {
		err := processModifySection(pkgbase, pconfig.Overrides.ModifySection)

		if err != nil {
			return err
		}
	}

	// Then run functions that merely append to the PKGBUILD

	if pconfig.Overrides.BumpEpoch > 0 {
		err := processBumpEpoch(pconfig, pkgbase)

		if err != nil {
			return err
		}
	}

	if pconfig.Overrides.BumpPkgrel != nil {
		err := processBumpPkgrel(pconfig, pkgbase)

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

	if pconfig.Overrides.ClearSignatures || pconfig.Overrides.RemoveSource != nil {
		err := processRemoveSources(pkgbase, pconfig.Overrides)

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

func (pconfig *PackageConfig) ProcessVcsOverrides(pkgbase string) error {
	if pconfig.Vcs == nil {
		return nil
	}

	if pconfig.Vcs.SourceOverrides != nil {
		err := processVcsSrcOverrides(pkgbase, pconfig.Vcs.SourceOverrides)

		if err != nil {
			return err
		}
	}

	modifySections := []*PackageConfigModifySection{
		{
			Type:    "function",
			Section: "pkgver",
			Replace: []*PackageConfigOverrideFromTo{
				{
					From: "(?m)^.*$",
				},
			},
		},
		{
			Type:    "variable",
			Section: "pkgver",
			Replace: []*PackageConfigOverrideFromTo{
				{
					From: "^.*$",
					To:   pconfig.Vcs.Pkgver,
				},
			},
		},
		{
			Type:    "variable",
			Section: "pkgrel",
			Replace: []*PackageConfigOverrideFromTo{
				{
					From: "^.*$",
					To:   strconv.Itoa(pconfig.Vcs.Pkgrel),
				},
			},
		},
	}

	err := processModifySection(pkgbase, modifySections)

	if err != nil {
		return err
	}

	return nil
}

func processBumpEpoch(pconfig *PackageConfig, pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing pkgrel bump overrides for pkgbase %s", pkgbase))

	tmpl := template.New("t")
	tmpl, err := tmpl.Parse(`
epoch=$((epoch + {{ .Bump }}))
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

	slog.Debug(fmt.Sprintf("Adding %d to epoch for pkgbase %s", pconfig.Overrides.BumpEpoch, pkgbase))

	err = tmpl.Execute(pkgbuild, map[string]string{
		"Bump": strconv.Itoa(pconfig.Overrides.BumpEpoch),
	})

	if err != nil {
		return err
	}

	return nil
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

	return nil
}

func processClearDependsVersions(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear depends versions override for pkgbase %s", pkgbase))

	appendText := `mapfile -t depends < <((IFS=$'\n'; echo "${depends[*]}") | sed -E 's/[<>=].*$//' | sort | uniq)`

	return appendPkgbuild(pkgbase, appendText)
}

func processRemoveSources(pkgbase string, overrides *PackageConfigOverrides) error {
	slog.Debug(fmt.Sprintf("Processing remove sources override for pkgbase %s", pkgbase))

	if overrides.ClearSignatures {
		if err := appendPkgbuild(pkgbase, "unset validpgpkeys"); err != nil {
			return err
		}
	}

	return removeSource(pkgbase, func(val string) (bool, error) {
		parts := strings.SplitN(val, "::", 2)
		filename := path.Base(parts[0])

		if overrides.ClearSignatures {
			if isSig, err := isSignature(filename); err != nil {
				return false, err
			} else if isSig {
				return true, nil
			}
		}

		if overrides.RemoveSource != nil {
			for _, source := range overrides.RemoveSource {
				if matched, err := regexp.MatchString(source, filename); err != nil {
					return false, err
				} else if matched {
					return true, nil
				}
			}
		}

		return false, nil
	})
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

func processModifySection(pkgbase string, overrides []*PackageConfigModifySection) error {
	slog.Debug(fmt.Sprintf("Processing modify section overrides for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuildBytes, err := os.ReadFile(pkgbuildPath)

	if err != nil {
		return err
	}

	pkgbuildLines := strings.Split(string(pkgbuildBytes), "\n")

	for _, override := range overrides {
		sections := override.Sections
		packages := override.Packages

		if override.Section != "" {
			sections = append(sections, override.Section)
		}

		if override.Package != "" {
			packages = append(packages, override.Package)
		}

		if len(sections) == 0 {
			sections = append(sections, "")
		}

		if len(packages) == 0 {
			packages = append(packages, "")
		}

		for _, sectionName := range sections {
			renamedSectionName := sectionName

			if override.Rename != "" {
				renamedSectionName = override.Rename
			}

			for _, packageName := range packages {
				if sectionName == "" && packageName != "" {
					return errors.New("cannot specify package name without section name")
				}

				beforeLines := []string{}
				afterLines := []string{}
				sectionLines := []string{}

				if packageName != "" {
					sectionName = fmt.Sprintf("%s_%s", sectionName, packageName)
					renamedSectionName = fmt.Sprintf("%s_%s", renamedSectionName, packageName)
				}

				functionStartLine := fmt.Sprintf("%s() {", sectionName)
				arrayStartLine := fmt.Sprintf("%s=(", sectionName)
				variableStartLine := fmt.Sprintf("%s=", sectionName)

				renamedFunctionStartLine := fmt.Sprintf("%s() {", renamedSectionName)
				renamedArrayStartLine := fmt.Sprintf("%s=(", renamedSectionName)
				renamedVariableStartLine := fmt.Sprintf("%s=", renamedSectionName)

				if sectionName != "" {
					foundStart := false
					foundEnd := false

					for _, line := range pkgbuildLines {
						if !foundStart {
							if (override.Type == "" || override.Type == "function") && line == functionStartLine {
								override.Type = "function"

								foundStart = true
								line = renamedFunctionStartLine
							} else if (override.Type == "" || override.Type == "array") && strings.HasPrefix(line, arrayStartLine) {
								override.Type = "array"

								foundStart = true
								foundEnd = true
								line = renamedArrayStartLine

								sectionLines = append(sectionLines, line)
								continue
							} else if (override.Type == "" || override.Type == "variable") && strings.HasPrefix(line, variableStartLine) {
								override.Type = "variable"

								foundStart = true
								foundEnd = true
								line = renamedVariableStartLine

								sectionLines = append(sectionLines, line)
								continue
							}

							beforeLines = append(beforeLines, line)
							continue
						}

						if !foundEnd {
							if line == "}" {
								foundEnd = true
							} else {
								sectionLines = append(sectionLines, line)
								continue
							}
						}

						afterLines = append(afterLines, line)
					}

					if !foundStart || !foundEnd {
						if override.Type == "" {
							if override.Append == "" && override.Prepend == "" {
								break
							}

							if !foundStart {
								return errors.New(fmt.Sprintf("Could not find start of %s section. Please specify Type field to create section.", sectionName))
							} else if !foundEnd {
								return errors.New(fmt.Sprintf("Could not find end of %s section. Please specify Type field to create section.", sectionName))
							}
						} else if override.Type == "function" {
							beforeLines = append(beforeLines, renamedFunctionStartLine)
							afterLines = append(afterLines, "}")
						} else if override.Type == "array" {
							sectionLines = append(sectionLines, fmt.Sprintf("%s)", renamedArrayStartLine))
						} else if override.Type == "variable" {
							sectionLines = append(sectionLines, renamedVariableStartLine)
						}
					}
				} else {
					sectionLines = pkgbuildLines[:]
				}

				if override.Type == "function" {
					if len(override.Replace) > 0 {
						sectionStr := strings.Join(sectionLines, "\n")

						for _, item := range override.Replace {
							re, err := regexp.Compile(item.From)

							if err != nil {
								return err
							}

							sectionStr = re.ReplaceAllString(sectionStr, item.To)
						}

						sectionLines = misc.FilterEmptyString(strings.Split(sectionStr, "\n"))
					}

					if len(override.Prepend) > 0 {
						sectionLines = append(strings.Split(override.Prepend, "\n"), sectionLines...)
					}

					if len(override.Append) > 0 {
						sectionLines = append(sectionLines, strings.Split(override.Append, "\n")...)
					}

					if len(sectionLines) == 0 {
						beforeLines = beforeLines[:len(beforeLines)-1]
						afterLines = afterLines[1:]
					}
				} else if override.Type == "array" {
					_, sectionItems, err := arrayLineToItems(sectionLines[0])

					if err != nil {
						return err
					}

					if len(override.Replace) > 0 {
						for _, item := range override.Replace {
							re, err := regexp.Compile(item.From)

							if err != nil {
								return err
							}

							for sItemIndex, sItem := range sectionItems {
								sectionItems[sItemIndex] = re.ReplaceAllString(sItem, item.To)
							}
						}
					}

					sectionItems = append(strings.Split(override.Prepend, "\n"), sectionItems...)
					sectionItems = append(sectionItems, strings.Split(override.Append, "\n")...)
					joinedSectionItems := strings.Join(sectionItems, " ")

					if len(strings.Trim(joinedSectionItems, " ")) > 0 {
						sectionLines[0] = fmt.Sprintf("%s%s)", renamedArrayStartLine, joinedSectionItems)
					} else {
						sectionLines = []string{}
					}
				} else if override.Type == "variable" {
					_, sectionValue, err := splitVariableLine(sectionLines[0])

					if err != nil {
						return err
					}

					if len(override.Replace) > 0 {
						for _, item := range override.Replace {
							re, err := regexp.Compile(item.From)

							if err != nil {
								return err
							}

							sectionValue = re.ReplaceAllString(sectionValue, item.To)
						}
					}

					sectionValue = fmt.Sprintf("%s%s%s", override.Prepend, sectionValue, override.Append)

					if len(strings.Trim(sectionValue, " ")) > 0 {
						sectionLines[0] = fmt.Sprintf("%s%s", renamedVariableStartLine, sectionValue)
					} else {
						sectionLines = []string{}
					}
				}

				pkgbuildLines = append(beforeLines, sectionLines...)
				pkgbuildLines = append(pkgbuildLines, afterLines...)
			}
		}
	}

	if err := os.WriteFile(pkgbuildPath, []byte(strings.Join(pkgbuildLines, "\n")), 0666); err != nil {
		return err
	}

	return nil
}

func processRenameFile(pkgbase string, overrides []*PackageConfigOverrideFromTo) error {
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

func processRenamePackage(pkgbase string, overrides []*PackageConfigOverrideFromTo) error {
	slog.Debug(fmt.Sprintf("Processing rename package override for pkgbase %s", pkgbase))

	packages, err := pacman.GetPkgbuildVars(pkgbase, "pkgname")

	if err != nil {
		return err
	}

	var pkgnames []string
	namechangemap := map[string]string{}
	functypenames := []string{"package", "prepare", "build", "check"}

	for _, pkgname := range packages {
		found := false

		for _, override := range overrides {
			if override.From == pkgname || (override.From == "" && pkgname == pkgbase) {
				if override.To != "" {
					pkgnames = append(pkgnames, override.To)

					for _, functypename := range functypenames {
						namechangemap[fmt.Sprintf("%s_%s", functypename, pkgname)] = fmt.Sprintf("%s_%s", functypename, override.To)
					}
				}

				found = true
				break
			}
		}

		if !found {
			pkgnames = append(pkgnames, pkgname)
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

func isSignature(value string) (bool, error) {
	parts := strings.SplitN(value, "::", 2)
	sigExtensions := []string{
		".sig",
		".sign",
		".asc",
	}

	for _, ext := range sigExtensions {
		if strings.HasSuffix(parts[0], ext) {
			return true, nil
		}
	}

	return false, nil
}

func removeSource(pkgbase string, matcher func(string) (bool, error)) error {
	slog.Debug(fmt.Sprintf("Removing source files for pkgbase %s", pkgbase))

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

	var appendLines []string
	var affectedArrays []string

	for _, item := range parsedArrays {
		if item.Name != "source" {
			continue
		}

		for index, value := range item.Items {
			if remove, err := matcher(value); err != nil {
				return err
			} else if !remove {
				continue
			}

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

	for _, item := range affectedArrays {
		appendLines = append(appendLines, fmt.Sprintf(`mapfile -t %s < <(IFS=$'\n'; echo "${%s[*]}")`, item, item))
	}

	appendText := strings.Join(appendLines, "\n")

	return appendPkgbuild(pkgbase, appendText)
}

func arrayLineToItems(line string) (string, []string, error) {
	parts := strings.SplitN(line, "=", 2)

	if len(parts) != 2 {
		return "", nil, errors.New("Invalid array line.")
	}

	sectionName := parts[0]
	varsStr := strings.TrimPrefix(strings.TrimSuffix(parts[1], ")"), "(")

	re, err := regexp.Compile(`"[^"]+"|'[^']+'|[^ ]+`)

	if err != nil {
		return "", nil, err
	}

	results := re.FindAllString(varsStr, -1)

	return sectionName, results, nil
}

func splitVariableLine(line string) (string, string, error) {
	parts := strings.SplitN(line, "=", 2)

	if len(parts) != 2 {
		return "", "", errors.New("Invalid variable line.")
	}

	return parts[0], parts[1], nil
}

func processVcsSrcOverrides(pkgbase string, overrides []*PackageConfigOverrideFromTo) error {
	slog.Debug(fmt.Sprintf("Processing vcs source overrides %s", pkgbase))

	appendText := `
mapfile -t _SOURCE_ARRAYS < <(compgen -v source)

for _SOURCE_ARRAY in "${_SOURCE_ARRAYS[@]}"; do
    for ((_SOURCE_INDEX = 0; _SOURCE_INDEX < $(eval echo '"'"\${#${_SOURCE_ARRAY}[*]}"'"'); _SOURCE_INDEX++)); do
        _SOURCE_VALUE="$(eval echo '"'"\${${_SOURCE_ARRAY}[${_SOURCE_INDEX}]}"'"')"
%s
    done
done

unset _SOURCE_ARRAYS
unset _SOURCE_ARRAY
unset _SOURCE_INDEX
unset _SOURCE_VALUE
`

	ifText := `

        if [ "${_SOURCE_VALUE}" = "%s" ]; then
            eval ${_SOURCE_ARRAY}[${_SOURCE_INDEX}]='"'"%s"'"'
        fi
`

	ifSections := []string{}

	for _, override := range overrides {
		ifSections = append(ifSections, fmt.Sprintf(ifText, override.From, override.To))
	}

	appendText = fmt.Sprintf(appendText, strings.Join(ifSections, "\n"))

	return appendPkgbuild(pkgbase, appendText)
}
