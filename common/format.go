package common

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"path/filepath"
)

//Regex to remove tabs and newlines (const)
var tabsNewlines = regexp.MustCompile(`\n(\t)?`)

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
		//Replace newlines by #012 and remove indentations
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

//ReplaceNewlines replaces newlines with #012 and removes indentations
//Arguments: a string for newline replacement
//Returns: string with #012 instead of newlines
func ReplaceNewlines(msg string) string {
	return tabsNewlines.ReplaceAllString(msg, "#012")
}
