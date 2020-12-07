package bear

import "errors"

var (
	ErrEmptyTemplate = errors.New("bear: empty template")
	ErrNoDB          = errors.New("bear: no db")
	ErrMismatchArgs  = errors.New("bear: mismatch args")
)
