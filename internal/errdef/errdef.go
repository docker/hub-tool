package errdef

import "errors"

// ErrCanceled represents a normally canceled operation
var ErrCanceled = errors.New("canceled")