package bear

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	debug       = true
	logInstance = log.New(os.Stdout, "", log.LstdFlags)
)

func EnableDebug(ok bool) {
	debug = ok
}

func SetLogOutput(w io.Writer) {
	logInstance.SetOutput(w)
}

func debugf(format string, args ...interface{}) {
	if debug {
		logInstance.Print("bear: " + fmt.Sprintf(format, args...))
	}
}
