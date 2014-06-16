/*
Deprecated. See "github.com/rightscale/rlog/console"
*/
package stdout

import (
	"fmt"
	"github.com/rightscale/rlog/console"
)

// deprecated
func NewStdoutLogger(removeNewlines bool) *console.ConsoleLogger {
	fmt.Println("rlog.stdout.NewStdoutLogger() is deprecated and will be removed in a future version.\nUse rlog.console.NewStdoutLogger() instead.")
	return console.NewStdoutLogger(removeNewlines)
}
