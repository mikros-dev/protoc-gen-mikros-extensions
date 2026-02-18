package log

import (
	"fmt"
	"io"
	"os"
)

const (
	prefix = "[mikros-extensions] "
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
	prefix string
}

func (d diagnosticLogger) Printf(format string, v ...any) {
	_, _ = fmt.Fprintf(d.writer, d.prefix+" "+format, v...)
}

func (d diagnosticLogger) Println(v ...any) {
	s := fmt.Sprintln(v...)
	_, _ = fmt.Fprint(d.writer, d.prefix+" "+s)
}

// LoggerOptions defines the options for a Logger.
type LoggerOptions struct {
	Verbose bool
	Prefix  string
}

// New returns a Logger. If verbose is false, it returns a logger that does
// nothing.
func New(options LoggerOptions) Logger {
	if !options.Verbose {
		return noopLogger{}
	}

	if options.Prefix == "" {
		options.Prefix = prefix
	}

	return diagnosticLogger{
		writer: os.Stderr,
		prefix: options.Prefix,
	}
}
