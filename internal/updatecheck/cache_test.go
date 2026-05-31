package updatecheck_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/buildoutinc/injector/internal/updatecheck"
)

func TestCache_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := updatecheck.Record{
		CheckedAt:     time.Now().UTC().Truncate(time.Second),
		LatestVersion: "v0.3.0",
		BinaryVersion: "v0.1.0",
	}
	if err := updatecheck.WriteCache(dir, in); err != nil {
		t.Fatalf("WriteCache: %v", err)
	}
	out, ok, err := updatecheck.ReadCache(dir)
	if err != nil {
		t.Fatalf("ReadCache: %v", err)
	}
	if !ok {
		t.Fatal("expected cache present")
	}
	if !out.CheckedAt.Equal(in.CheckedAt) || out.LatestVersion != in.LatestVersion || out.BinaryVersion != in.BinaryVersion {
		t.Fatalf("round-trip mismatch: in=%+v out=%+v", in, out)
	}
}

func TestCache_MissingFileIsCacheMiss(t *testing.T) {
	dir := t.TempDir()
	_, ok, err := updatecheck.ReadCache(dir)
	if err != nil {
		t.Fatalf("err on missing: %v", err)
	}
	if ok {
		t.Fatal("expected miss")
	}
}

func TestCache_MalformedReturnsCacheMiss(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "update-check.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, ok, err := updatecheck.ReadCache(dir)
	if err != nil {
		t.Fatalf("err on garbage: %v", err)
	}
	if ok {
		t.Fatal("expected garbage to be treated as miss")
	}
}
