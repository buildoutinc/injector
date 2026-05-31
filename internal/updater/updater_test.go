package updater_test

import (
	"context"
	"errors"
	"testing"

	"github.com/buildoutinc/injector/internal/updater"
)

func TestFakeUpdater_RoundTrip(t *testing.T) {
	target := updater.Release{Version: "v0.2.0", AssetURL: "http://x"}
	want := updater.Result{FromVersion: "v0.1.0", ToVersion: "v0.2.0", Path: "/tmp/inject"}

	f := &updater.FakeUpdater{
		LatestFunc: func(ctx context.Context, opts updater.LatestOpts) (updater.Release, error) {
			if !opts.IncludePrereleases {
				return updater.Release{}, errors.New("flag not propagated")
			}
			return target, nil
		},
		ApplyFunc: func(ctx context.Context, r updater.Release, p string) (updater.Result, error) {
			if r != target || p != "/tmp/inject" {
				t.Fatalf("apply got (%+v, %q)", r, p)
			}
			return want, nil
		},
	}

	got, err := f.Latest(context.Background(), updater.LatestOpts{IncludePrereleases: true})
	if err != nil || got != target {
		t.Fatalf("Latest = (%+v, %v); want %+v", got, err, target)
	}
	res, err := f.Apply(context.Background(), got, "/tmp/inject")
	if err != nil || res != want {
		t.Fatalf("Apply = (%+v, %v); want %+v", res, err, want)
	}
	if f.LatestCallCount() != 1 || f.ApplyCallCount() != 1 {
		t.Fatalf("call counts: latest=%d apply=%d", f.LatestCallCount(), f.ApplyCallCount())
	}
}
