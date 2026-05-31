package updater

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// ErrNotWritable indicates the binary path is not writable by the current user.
// Carries the offending path so the CLI can format a remediation message.
type ErrNotWritable struct {
	Path string
	Err  error
}

func (e *ErrNotWritable) Error() string {
	return fmt.Sprintf("%s is not writable by your user: %v", e.Path, e.Err)
}

func (e *ErrNotWritable) Unwrap() error { return e.Err }

// IsNotWritable reports whether err is an ErrNotWritable.
func IsNotWritable(err error) bool {
	var nw *ErrNotWritable
	return errors.As(err, &nw)
}

// CheckWritable returns ErrNotWritable when the current process cannot
// write to path. Uses access(2) so ACLs and read-only mounts are honored.
func CheckWritable(path string) error {
	if err := unix.Access(path, unix.W_OK); err != nil {
		return &ErrNotWritable{Path: path, Err: err}
	}
	return nil
}
