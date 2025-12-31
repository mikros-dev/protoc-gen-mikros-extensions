package log

import (
	"fmt"
	"io"
	"os"
)

// Logger defines the interface for plugin logging.
type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

// noopLogger discards all logs.
type noopLogger struct{}

func (n noopLogger) Printf(format string, v ...any) {
	// noop
}

func (n noopLogger) Println(v ...any) {
	// noop
}

// diagnosticLogger writes to a specific io.Writer
type diagnosticLogger struct {
	writer io.Writer
}

func (d diagnosticLogger) Printf(format string, v ...any) {
	_, _ = fmt.Fprintf(d.writer, format, v...)
}

func (d diagnosticLogger) Println(v ...any) {
	_, _ = fmt.Fprintln(d.writer, v...)
}

// New returns a Logger. If verbose is false, it returns a logger that does
// nothing.
func New(verbose bool) Logger {
	if !verbose {
		return noopLogger{}
	}

	return diagnosticLogger{
		writer: os.Stderr,
	}
}
