package bear

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	debug       = true
	logPrefix   = "bear: "
	logInstance = log.New(os.Stdout, "", log.LstdFlags)
	errPrefix   = "bear: "
)

func EnableDebug(ok bool) {
	debug = ok
}

func SetLogOutput(w io.Writer) {
	logInstance.SetOutput(w)
}

func debugf(format string, args ...interface{}) {
	if debug {
		logInstance.Print(logPrefix + fmt.Sprintf(format, args...))
	}
}

func errorf(format string, args ...interface{}) error {
	return errors.New("bear:" + fmt.Sprintf(format, args...))
}
