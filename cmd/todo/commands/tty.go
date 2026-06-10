package commands

import (
	"io"
	"os"
)

var isTerminal = func(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

const (
	ansiRed   = "\x1b[31m"
	ansiReset = "\x1b[0m"
)
