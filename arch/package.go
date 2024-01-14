package arch

import (
	"database/sql"
	"fmt"
	"github.com/ryanpetris/aur-builder/pacdb"
	"log/slog"
	"strings"
)

type Package struct {
	Pkgbase string `pacdb:"base"`
	Pkgname string `pacdb:"package"`
	Version string `pacdb:"version"`
}

func PackageExists(pkgbase string) (bool, error) {
	query := "SELECT COUNT(*) FROM packages WHERE base = :base"
	params := []any{
		sql.Named("base", pkgbase),
	}

	var count int

	if err := pacdb.QueryRow(query, params, &count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetPackages(pkgnames []string) ([]Package, error) {
	if len(pkgnames) == 0 {
		return nil, nil
	}

	var paramNames []string
	var params []any

	for i, pkgname := range pkgnames {
		paramName := fmt.Sprintf("package%d", i)
		paramNames = append(paramNames, fmt.Sprintf(":%s", paramName))
		params = append(params, sql.Named(paramName, pkgname))
	}

	query := fmt.Sprintf("SELECT base, package, version FROM packages WHERE db IN ('core', 'extra') AND package IN (%s)", strings.Join(paramNames, ", "))

	return pacdb.QueryStruct[Package](query, params)
}

func GetPackageVersion(pkgname string) (string, error) {
	slog.Debug(fmt.Sprintf("Looking up version for package %s", pkgname))

	query := "SELECT version FROM packages WHERE package = :package"
	params := []any{
		sql.Named("package", pkgname),
	}

	var version string

	if err := pacdb.QueryRow(query, params, &version); err != nil {
		return "", err
	} else if version != "" {
		return version, nil
	}

	slog.Debug(fmt.Sprintf("Version not found in output for package %s", pkgname))

	return "", nil
}
