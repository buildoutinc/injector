package updater

import (
	"context"
	"errors"
	"sync"
)

// FakeUpdater is an in-memory Updater for tests. All behavior is
// configured via the *Func fields; nil funcs return zero values.
type FakeUpdater struct {
	mu sync.Mutex

	LatestFunc func(ctx context.Context, opts LatestOpts) (Release, error)
	ApplyFunc  func(ctx context.Context, target Release, binPath string) (Result, error)

	LatestCalls []LatestOpts
	ApplyCalls  []FakeApplyCall
}

type FakeApplyCall struct {
	Target  Release
	BinPath string
}

func (f *FakeUpdater) Latest(ctx context.Context, opts LatestOpts) (Release, error) {
	f.mu.Lock()
	f.LatestCalls = append(f.LatestCalls, opts)
	fn := f.LatestFunc
	f.mu.Unlock()
	if fn == nil {
		return Release{}, errors.New("FakeUpdater.LatestFunc not set")
	}
	return fn(ctx, opts)
}

func (f *FakeUpdater) Apply(ctx context.Context, target Release, binPath string) (Result, error) {
	f.mu.Lock()
	f.ApplyCalls = append(f.ApplyCalls, FakeApplyCall{Target: target, BinPath: binPath})
	fn := f.ApplyFunc
	f.mu.Unlock()
	if fn == nil {
		return Result{}, errors.New("FakeUpdater.ApplyFunc not set")
	}
	return fn(ctx, target, binPath)
}

func (f *FakeUpdater) LatestCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.LatestCalls)
}

func (f *FakeUpdater) ApplyCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.ApplyCalls)
}
