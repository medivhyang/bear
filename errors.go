package bear

import "errors"

var (
	ErrEmptyTemplate = errors.New("bear: empty template")
	ErrRequireDB     = errors.New("bear: require db")
	ErrInvalidArgs   = errors.New("bear: invalid args")
)
