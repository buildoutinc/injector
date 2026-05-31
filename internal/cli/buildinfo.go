package cli

var (
	pkgVersion = "dev"
	pkgCommit  = "none"
	pkgDate    = "unknown"
)

// SetBuildInfo is called once by main with the linker-injected values.
func SetBuildInfo(version, commit, date string) {
	pkgVersion = version
	pkgCommit = commit
	pkgDate = date
}

// Version returns the linker-injected semver (or "dev" for local builds).
func Version() string { return pkgVersion }

// BuildInfo returns the commit SHA and build date baked at link time.
func BuildInfo() (commit, date string) { return pkgCommit, pkgDate }
