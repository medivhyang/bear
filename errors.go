package bear

import "errors"

var (
	ErrEmptyTemplate   = errors.New("bear: empty template")
	ErrRequireDB       = errors.New("bear: require db")
	ErrRequireExecutor = errors.New("bear: require executor")
)
