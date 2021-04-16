package bear

import (
	"fmt"
	"io"
)

var (
	Debug     = false
	LogWriter io.Writer
)

func debugf(format string, args ...interface{}) {
	if Debug {
		if _, err := fmt.Fprintf(LogWriter, format, args...); err != nil {
			panic(err)
		}
	}
}
