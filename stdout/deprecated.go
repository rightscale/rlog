/*
Deprecated. See "github.com/brsc/rlog/console"
*/
package stdout

import (
  "github.com/brsc/rlog/console"
)

// deprecated
func NewStdoutLogger(removeNewlines bool) *console.ConsoleLogger {
  return console.NewStdoutLogger(removeNewlines)
}
