// Package updater is the boundary around the self-update library.
// CLI code talks to the Updater interface so unit tests can substitute
// an in-memory fake (Constitution VI: tests run offline).
package updater

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// Updater is the small surface the CLI uses to discover and apply releases.
type Updater interface {
	Latest(ctx context.Context, opts LatestOpts) (Release, error)
	Apply(ctx context.Context, target Release, binPath string) (Result, error)
}

// LatestOpts shapes the discovery request.
type LatestOpts struct {
	// IncludePrereleases asks the updater to consider prerelease tags.
	// NOTE: the underlying go-github-selfupdate v1.2.3 always skips
	// prereleases on its DetectLatest path; this flag is plumbed for
	// forward-compatibility and is observed by FakeUpdater.
	IncludePrereleases bool
}

// Release describes a single release candidate.
type Release struct {
	Version  string // "vX.Y.Z"
	URL      string // human-facing release URL
	AssetURL string // direct URL of the OS/arch archive
}

// Result is what Apply returns on a successful swap.
type Result struct {
	FromVersion string
	ToVersion   string
	Path        string // the binary that was replaced
}

// Sentinel errors for the CLI to recognize.
var (
	ErrNoAssetForPlatform = errors.New("no release artifact for current OS/arch")
	ErrChecksumMismatch   = errors.New("checksum mismatch")
)

// githubUpdater is the production Updater backed by go-github-selfupdate.
type githubUpdater struct {
	slug           string
	currentVersion string
}

// NewGithub constructs a real Updater pointing at the named GitHub repo
// ("owner/name"). The currentVersion must be a parsable semver (with or
// without a leading "v"); "dev" and other non-semver strings cause Apply
// to refuse and Latest to behave as if no release exists.
func NewGithub(slug, currentVersion string) Updater {
	return &githubUpdater{slug: slug, currentVersion: currentVersion}
}

func (u *githubUpdater) Latest(ctx context.Context, _ LatestOpts) (Release, error) {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.SHA2Validator{},
	})
	if err != nil {
		return Release{}, fmt.Errorf("init updater: %w", err)
	}
	rel, found, err := updater.DetectLatest(u.slug)
	if err != nil {
		return Release{}, err
	}
	if !found || rel == nil {
		return Release{}, fmt.Errorf("%w: %s/%s", ErrNoAssetForPlatform, runtime.GOOS, runtime.GOARCH)
	}
	return Release{
		Version:  "v" + rel.Version.String(),
		URL:      rel.URL,
		AssetURL: rel.AssetURL,
	}, nil
}

func (u *githubUpdater) Apply(ctx context.Context, target Release, binPath string) (Result, error) {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.SHA2Validator{},
	})
	if err != nil {
		return Result{}, fmt.Errorf("init updater: %w", err)
	}
	v, err := semver.ParseTolerant(target.Version)
	if err != nil {
		return Result{}, fmt.Errorf("parse target version: %w", err)
	}
	rel := &selfupdate.Release{
		Version:  v,
		AssetURL: target.AssetURL,
		URL:      target.URL,
		RepoOwner: ownerOf(u.slug),
		RepoName:  nameOf(u.slug),
	}
	if err := updater.UpdateTo(rel, binPath); err != nil {
		if errMentionsChecksum(err) {
			return Result{}, fmt.Errorf("%w: %v", ErrChecksumMismatch, err)
		}
		return Result{}, err
	}
	return Result{FromVersion: u.currentVersion, ToVersion: target.Version, Path: binPath}, nil
}

func ownerOf(slug string) string {
	for i := 0; i < len(slug); i++ {
		if slug[i] == '/' {
			return slug[:i]
		}
	}
	return slug
}

func nameOf(slug string) string {
	for i := 0; i < len(slug); i++ {
		if slug[i] == '/' {
			return slug[i+1:]
		}
	}
	return ""
}

func errMentionsChecksum(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, needle := range []string{"checksum", "hash mismatch", "sha256"} {
		if containsFold(msg, needle) {
			return true
		}
	}
	return false
}

func containsFold(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			a, b := s[i+j], substr[j]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
