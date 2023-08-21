package misc

type PackageTracker struct {
	Pkgbase           string
	UpstreamVersion   string
	RepositoryVersion string
	NeedsUpdate       bool
	Packages          []PackageInfo
}
