package pkg

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"log/slog"
	"os"
	"path"
	"strconv"
	"text/template"
)

func (pconfig *PackageConfig) ProcessOverrides(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing overrides for pkgbase %s", pkgbase))

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

	if pconfig.Overrides.ClearPkgverFunc {
		err := processClearPkgverFunc(pkgbase)

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

	if err != nil {
		return err
	}

	return nil
}

func processClearDependsVersions(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear depends versions override for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuild, err := os.OpenFile(pkgbuildPath, os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		return err
	}

	defer pkgbuild.Close()

	appendText := `mapfile -t depends < <((IFS=$'\n'; echo "${depends[*]}") | sed -E 's/[<>=].*$//' | sort | uniq)`

	if _, err = pkgbuild.WriteString(appendText); err != nil {
		return err
	}

	return nil
}

func processClearPkgverFunc(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear pkgver function override for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuild, err := os.OpenFile(pkgbuildPath, os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		return err
	}

	defer pkgbuild.Close()

	appendText := `unset -f pkgver`

	if _, err = pkgbuild.WriteString(appendText); err != nil {
		return err
	}

	return nil
}

func processClearSignatures(pkgbase string) error {
	slog.Debug(fmt.Sprintf("Processing clear signatures override for pkgbase %s", pkgbase))

	mergedPath := config.GetMergedPath(pkgbase)
	pkgbuildPath := path.Join(mergedPath, "PKGBUILD")
	pkgbuild, err := os.OpenFile(pkgbuildPath, os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		return err
	}

	defer pkgbuild.Close()

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

	if _, err = pkgbuild.WriteString(appendText); err != nil {
		return err
	}

	return nil
}
