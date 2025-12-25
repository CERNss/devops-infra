package executor

import (
	"fmt"
	"os"
	"strings"
)

func Prepare(cmd string, opts Options) string {
	if opts.Sudo && !isRoot() {
		return "sudo -E bash -c " + shellQuote(cmd)
	}
	return cmd
}

func isRoot() bool {
	return os.Geteuid() == 0
}

func shellQuote(cmd string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(cmd, `'`, `'\''`))
}
