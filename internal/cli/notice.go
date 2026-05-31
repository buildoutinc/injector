package cli

import (
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/term"

	"github.com/buildoutinc/injector/internal/updatecheck"
)

var noticeOnce sync.Once

// RenderNotice writes the two-line `==>` upgrade nudge to w, or nothing
// when n is nil. Guarded by a process-wide sync.Once so the notice
// surfaces at most once per command invocation even if multiple paths
// (help screen + subcommand) both call it.
func RenderNotice(w io.Writer, n *updatecheck.Notice, tty bool) {
	if n == nil {
		return
	}
	noticeOnce.Do(func() {
		marker := "==>"
		if tty {
			marker = "\x1b[36m==>\x1b[0m"
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s A newer version is available: %s\n", marker, n.Latest)
		fmt.Fprintf(w, "%s Run \"inject upgrade\" to install it.\n", marker)
	})
}

// resetNoticeOnce is called by tests; safe to leave a no-op in
// production because main only runs once.
func resetNoticeOnce() { noticeOnce = sync.Once{} }

// isTTY reports whether w is a terminal. Returns false for buffers etc.
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
