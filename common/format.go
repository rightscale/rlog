package common

import (
  "fmt"
  "os"
  "path/filepath"
  "regexp"
  "strconv"
  "strings"
)

// Regex to remove tabs and newlines.
const (
  replacementWhitespacePattern = `[\r\n\t]+`
  replacementWhitespace        = "  "
)

var replaceWhitespaceRegex = regexp.MustCompile(replacementWhitespacePattern)

//SyslogHeader gathers environment information to generate a log prefix
func SyslogHeader() string {
  //Fetch process name, pid and hostname
  _, processName := filepath.Split(os.Args[0])
  pid := strconv.Itoa(os.Getpid())
  hostname, err := os.Hostname()

  if err != nil {
    //This is a non-fatal error and hence we just print a message
    fmt.Printf("rlog initialization error: could not fetch machine hostname")
  }

  //Generate a prefix out of this information
  prefix := hostname + " " + processName + "[" + pid + "]: "

  return prefix
}

//FormatMessage generates a log message
func FormatMessage(rawRlogMsg *RlogMsg, prefix string, removeNewlines bool) string {
  logMsg := rawRlogMsg.Msg
  trace := rawRlogMsg.StackTrace
  if removeNewlines {
    logMsg = ReplaceNewlines(logMsg)
  }

  //Print the log message and stack trace if appropriate
  res := rawRlogMsg.Timestamp + " " + prefix + logMsg
  if trace != "" {
    if removeNewlines {
      trace = ReplaceNewlines(trace)
      res += ", trace: " + trace
    } else {
      res += "\n" + trace
    }
  }

  return res
}

// Replaces any tabs/newlines with double-space and removes indentations due to
// newlines being displayed as #012 in syslog (which is ugly).
//
// msg: message for newline replacement
//
// return: message with newlines replaced
func ReplaceNewlines(msg string) string {
  return strings.Trim(
    replaceWhitespaceRegex.ReplaceAllString(msg, replacementWhitespace), " ")
}
