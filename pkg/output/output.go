package output

import (
	"fmt"
	"os"
)

var (
	enabled = false
)

func Enable(enable bool) {
	enabled = enable
}

func Println(values ...interface{}) {
	if enabled {
		_, _ = fmt.Fprint(os.Stderr, values...)
		println()
	}
}
