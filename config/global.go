package config

import (
	"sync"
)

var config *Config
var once sync.Once

func GetGlobalConfig() *Config {
	once.Do(func() {
		config = &Config{}
	})

	return config
}

func GetBasePath() string {
	config := GetGlobalConfig()

	return config.GetBasePath()
}

func GetPackagePath(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetPackagePath(pkgbase)
}

func GetConfigPath(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetConfigPath(pkgbase)
}

func GetLocalPath(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetLocalPath(pkgbase)
}

func GetMergedPath(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetMergedPath(pkgbase)
}

func GetScriptsPath(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetScriptsPath(pkgbase)
}

func GetScriptOverridePath(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetScriptOverridePath(pkgbase)
}

func GetUpstreamPath(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetUpstreamPath(pkgbase)
}

func GetAurBaseUrl() string {
	config := GetGlobalConfig()

	return config.GetAurBaseUrl()
}

func GetAurPackagesUrl() string {
	config := GetGlobalConfig()

	return config.GetAurPackagesUrl()
}

func GetAurPackageGitUrl(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetAurPackageGitUrl(pkgbase)
}

func GetArchBaseGitUrl() string {
	config := GetGlobalConfig()

	return config.GetArchBaseGitUrl()
}

func GetArchPackageGitUrl(pkgbase string) string {
	config := GetGlobalConfig()

	return config.GetArchPackageGitUrl(pkgbase)
}
