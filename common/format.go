package common

import (
	"fmt"
	"os"
	"path"
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
	processName := path.Base(os.Args[0])
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
		//Replace whitespace
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

//ReplaceNewlines any tabs/newlines with double-space and removes indentations
//Arguments: a string for newline replacement
//Returns: string with #012 instead of newlines
func ReplaceNewlines(msg string) string {
	return strings.Trim(
		replaceWhitespaceRegex.ReplaceAllString(msg, replacementWhitespace), " ")
}
