package diffsql

import "errors"

var (
	ErrDiffFailed            = errors.New("ErrDiffFailed")
	ErrDiffResultCheckFailed = errors.New("ErrDiffResultCheckFailed")
)
