/*
Package stdout implements an output module for logging to stdout using rlog.
*/
package stdout

import (
	"fmt"
	"github.com/rightscale/rlog/common"
)

//Configuration of stdout module
type stdoutModuleConfig struct {
	removeNewlines bool
}

//NewStdoutLogger enables logging to standard output (console)
//Arguments: remove newline flag, when set to true newlines are replaces by #012 as
//when printing to syslog
//Returns: instace of stdout logger
func NewStdoutLogger(removeNewlines bool) *stdoutModuleConfig {
	stdoutConf := new(stdoutModuleConfig)
	stdoutConf.removeNewlines = removeNewlines
	return stdoutConf
}

//LaunchModule is intended to run in a separate goroutine. It prints log messages to stdout
//Arguments: [dataChan] Channel to receive log messages. [flushChan] Channel to receive flush command
func (conf *stdoutModuleConfig) LaunchModule(dataChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {

	prefix := common.SyslogHeader()

	//Wait forever on data and flush channel
	for {
		select {
		case logMsg := <-dataChan:
			//Received log message, print it
			conf.printMsg(logMsg, prefix)
		case ret := <-flushChan:
			//Flush and return success
			conf.flush(dataChan, prefix)
			ret <- true
		}
	}
}

//printMsg prints the message to stdout
//Arguments: [rawRlogMsg] log message received from channel, [prefix] log prefix
func (conf *stdoutModuleConfig) printMsg(rawRlogMsg *common.RlogMsg, prefix string) {
	fmt.Println(common.FormatMessage(rawRlogMsg, prefix, conf.removeNewlines))
}

//flush writes all pending log messages to stdout
//Arguments:[dataChan] data channel to access all pending messages, [prefix] log prefix
func (conf *stdoutModuleConfig) flush(dataChan <-chan (*common.RlogMsg), prefix string) {
	for {
		//Perform non blocking read until the channel is empty
		select {
		case logMsg := <-dataChan:
			conf.printMsg(logMsg, prefix)
		default:
			return
		}
	}
}
