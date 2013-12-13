package rlog

/*
This file implements the log message generation and formatting. The genericLogHandler takes the message from
 the corresponding rlog API call (e.g. Error, Info, etc.) and controls processing until the log message is
forwarded to the logmsg channel of each registered module.
*/

import (
	"fmt"
	"github.com/brsc/rlog/common"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//logPieces keeps all raw information about a log message for further processing (formatting, etc.)
type logPieces struct {
	level      string              //log level.
	msg        string              //log message
	severity   common.RlogSeverity //log severity
	posInfo    bool                //does the log message need to be accompanied by file and line number?
	file       string              //file where log message was generated
	line       int                 //line where log message was generated.
	pc         uint                //program counter position where log message was generated
	stackTrace string              //stack trace (for error and fatal only)
}

//genericLogHandler is called from various sources like info, error, errorT, etc. It gathers all the data
//and controls the log message processing until the log message is distributed to the registered modules.
//Arguments: [level]: log level as it should appear in the log output (INFO, ERROR, etc.).
//[tag]: log message tag (nil if no tag). [format and a]: printf formatted message. [severity]: log message
//severity. [posInfo]: True if log message should include file and line number
//Returns: false if the logger is not initialized, true otherwise
func genericLogHandler(level string, tag string, format string, a []interface{}, severity common.RlogSeverity, posInfo bool) bool {

	if !initialized {
		//Ensure that logger is initialized
		log.Printf("[ERROR] Logger not initialized, msg: "+format, a...)
		return false
	}

	if isFilteredSeverity(severity) || isFilteredTag(tag) {
		//Drop message
		return true
	}

	//Gather data: create a struct to hold the raw data and fill it
	logMsg := fmt.Sprintf(format, a...)
	pc, file, line := getLogCallPos()

	trace := ""
	if severity <= SeverityError {
		//Obtain stack trace only for fatal and error
		trace = getStackTrace()
	}

	raw := logPieces{level, logMsg, severity, posInfo, file, line, pc, trace}

	//Apply algorithm to create a nicely formatted log message as rlog message
	sysLogMsg := raw.generateLogMsg()

	//All processing completed, send log message to syslog
	pushToChannels(sysLogMsg)
	return true
}

//getStackTrace generates a stack trace
//Returns: stack trace
func getStackTrace() string {
	//Fetch stack, store in buffer (buffer size limited to 1KB) and convert it to string
	buf := make([]byte, 2048)
	n := runtime.Stack(buf, false)
	str := string(buf[0:n])

	//The stack trace is represented as lines (2 lines ==> 1 level in call hierarchy). Cut off the first
	//4 hierarchy levels because they are rlog internal calls.
	//With SplitAfterN, we split (on \n) the stack trace into cutLines substrings ([]string), where the
	//last substring 	//will be the unsplit remainder. By taking [cutLines-1], we select exactly that
	//unsplit remainder which corresponds to the remainder of the stack trace.
	cutLines := 8
	res := strings.SplitAfterN(str, "\n", cutLines)[cutLines-1]
	res = strings.TrimRight(res, "\n") // Remove trailing newline
	return res
}

//generateLogMsg generates the actual log message from raw log information
//Returns: RlogMsg ready to send to the modules
func (lp *logPieces) generateLogMsg() *common.RlogMsg {
	sysLogMsg := new(common.RlogMsg)

	//Add formatted log message to struct
	header := formatHeaders(lp.posInfo, lp.level, lp.file, lp.line)
	sysLogMsg.Msg = header + lp.msg

	//Set additional parameters
	sysLogMsg.Severity = lp.severity
	sysLogMsg.Pc = lp.pc
	sysLogMsg.StackTrace = lp.stackTrace
	sysLogMsg.Timestamp = time.Now().Format(time.Stamp)

	return sysLogMsg
}

//formatHeaders creates a log message header.
//Arguments: [posInfo] determines whether file and line number should be included. [level] represents the log level
//as string. [file] File causing log message. [line] Line number in file causing log message.
//Returns: Formatted header
func formatHeaders(posInfo bool, level string, file string, line int) string {

	var header string

	//Add log level
	header += "<" + level + "> "

	if posInfo {
		//Add file and line number to log message
		header += "[" + file + ":" + strconv.Itoa(line) + "] "
	}

	return header
}

//isFilteredSeverity determines whether the given log message shall be filtered because of
//the severity configuration
func isFilteredSeverity(severity common.RlogSeverity) bool {
	return severity > config.Severity
}

//isFilteredSeverity determines whether the given log message shall be filtered due to tag
//configuration. A nil argument represents no tag
func isFilteredTag(tag string) bool {

	filtered := false
	if config.tagsEnabledExcept != nil {
		filtered, _ = config.tagsEnabledExcept[tag]
	} else if config.tagsDisabledExcept != nil {
		filtered, _ = config.tagsDisabledExcept[tag]
		filtered = !filtered
	}

	return filtered
}

//getLogCallPos obtains information about the place of the rlog invocation.
//Returns: program counter (pc), file and line of rlog invocation
func getLogCallPos() (uint, string, int) {
	//Important: the information is fetched 3 levels up. Consider the following nested function call:
	//a(b(c(getLogPos()))). getLogCallPos returns the context from method call b because this is where
	//the user of rlog printed a message

	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		log.Printf("Could not fetch log position information")
		//Set values to unknown, do not print an error message as there is nothing we can do about it
		pc = 0
		file = "unknown"
		line = 0
	}

	return uint(pc), file, line
}
