package bear

import (
	"fmt"
	"io"
	"os"
)

var (
	debugFlag             = false
	debugWriter io.Writer = os.Stdout
	logPrefix             = "bear: "
)

func Debug(b bool) {
	debugFlag = b
}

func Output(writer io.Writer) {
	debugWriter = writer
}

func debugf(module string, format string, args ...interface{}) {
	if !debugFlag {
		return
	}
	if debugWriter == nil {
		return
	}
	s := fmt.Sprintf("%s%s: %s", logPrefix, module, fmt.Sprintf(format, args...))
	if _, err := fmt.Fprintln(debugWriter, s); err != nil {
		panic(err)
	}
}

func newError(module string, format string, args ...interface{}) error {
	return fmt.Errorf("%s%s: %s", logPrefix, module, fmt.Sprintf(format, args...))
}
