package aur

import (
	"fmt"
	"github.com/ryanpetris/aur-builder/config"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
	"sync"
)
import "net/http"

var packages *[]string
var packagesOnceErr *error
var packagesOnce sync.Once

func PackageExists(pkgbase string) (bool, error) {
	packagesOnce.Do(func() {
		response, err := http.Get(config.GetAurPackagesUrl())

		if err != nil {
			packagesOnceErr = &err
			return
		}

		data, err := io.ReadAll(response.Body)

		if err != nil {
			packagesOnceErr = &err
			return
		}

		dataStr := strings.Split(string(data), "\n")

		packages = &dataStr
	})

	if packagesOnceErr != nil {
		return false, *packagesOnceErr
	}

	for _, line := range *packages {
		if line == pkgbase {
			return true, nil
		}
	}

	return false, nil
}

func GetPackageInfos(pkgbase []string) ([]Package, error) {
	aurRoot := config.GetAurBaseUrl()
	rpcUrl := fmt.Sprintf("%s/rpc/v5/info", aurRoot)

	for index, pkg := range pkgbase {
		prefix := "&"

		if index == 0 {
			prefix = "?"
		}

		rpcUrl = fmt.Sprintf("%s%sarg[]=%s", rpcUrl, prefix, pkg)
	}

	response, err := http.Get(rpcUrl)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	result := PackageSearchResults{}

	err = yaml.Unmarshal(data, &result)

	if err != nil {
		return nil, err
	}

	return result.Results, nil
}
