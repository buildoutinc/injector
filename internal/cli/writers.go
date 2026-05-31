package cli

import "io"

// Stderr is a distinct named type used as a kong bind target so the
// stderr writer doesn't collide with the stdout `io.Writer` binding.
type Stderr struct{ io.Writer }
