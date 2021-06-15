package bear

import (
	"fmt"
	"io"
	"os"
)

var (
	debugFlag             = false
	debugWriter io.Writer = os.Stdout
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
	s := fmt.Sprintf("%s: %s: %s", "bear", module, fmt.Sprintf(format, args...))
	if _, err := fmt.Fprintln(debugWriter, s); err != nil {
		panic(err)
	}
}

func errorf(module string, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %s: %s", "deer", module, fmt.Sprintf(format, args...))
}

func repeatString(s string, n int) []string {
	r := make([]string, 0, n)
	for i := 0; i < n; i++ {
		r = append(r, s)
	}
	return r
}
