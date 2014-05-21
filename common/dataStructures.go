/*
Package common contains data structures and constants shared between rlog and its modules.
*/
package common

//RlogMsg carries a formatted log message including some additional information.
type RlogMsg struct {
  Msg        string       //log message
  Timestamp  string       //time of log generation (preformatted)
  Severity   RlogSeverity //log severity
  Pc         uint         //program counter position where log message was generated
  StackTrace string       //stack trace (for error and fatal only)
}

//RlogSeverity defines a type to represent severity levels for log messages
type RlogSeverity uint
