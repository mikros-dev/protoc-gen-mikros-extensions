package output

import (
	"fmt"
	"os"
)

var (
	enabled = false
)

// Enable enables or disables the output of the plugin.
func Enable(enable bool) {
	enabled = enable
}

// Println prints the given values to stderr.
func Println(values ...interface{}) {
	if enabled {
		_, _ = fmt.Fprint(os.Stderr, values...)
		println()
	}
}
