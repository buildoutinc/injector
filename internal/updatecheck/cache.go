// Package updatecheck owns the rate-limited "is there a newer release"
// check that drives the passive upgrade notice.
package updatecheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheFileName = "update-check.json"

// Record is the persisted shape of the cache file.
type Record struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	BinaryVersion string    `json:"binary_version"`
}

// DefaultCacheDir returns and ensures the per-user cache dir for inject.
func DefaultCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "inject")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// ReadCache reads the cache record from dir. The second return value is
// false when the file is missing or unparsable (treated as a miss, not
// an error, so callers don't have to special-case first runs).
func ReadCache(dir string) (Record, bool, error) {
	path := filepath.Join(dir, cacheFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Record{}, false, nil
		}
		return Record{}, false, err
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		// Corrupt cache → treat as a miss; the next check will rewrite it.
		return Record{}, false, nil
	}
	if rec.CheckedAt.IsZero() {
		return Record{}, false, nil
	}
	return rec, true, nil
}

// WriteCache atomically writes the record to dir (tmp + rename).
func WriteCache(dir string, r Record) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".update-check-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, cacheFileName)); err != nil {
		return fmt.Errorf("rename cache: %w", err)
	}
	return nil
}
