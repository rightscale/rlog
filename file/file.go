/*
Package file implements an output module for logging to a file using rlog.
*/
package file

import (
	"github.com/brsc/rlog"
	"github.com/brsc/rlog/common"
	"os"
)

//Configuration of file logging module
type fileLogger struct {
	removeNewlines bool
	fileHandle     *os.File
	loggedError    bool
}

//NewFileLogger enables logging to a file. The path (path/filename) can be specified either relative
//to the application directory or as full path (example: "myLog.txt"). When removeNewlines is set,
//newlines and tabs are replaced with ASCII characters as in syslog. If overwrite is set, the log
//file is overwritten each time the application is restarted. If disabled, logs are appended.
func NewFileLogger(path string, removeNewlines bool, overwrite bool) (*fileLogger, error) {

	//Define file handle
	var fh *os.File
	var err error

	if overwrite {
		fh, err = os.Create(path)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = os.Stat(path)
		if os.IsNotExist(err) {

			//File not present, create it
			fh, err = os.Create(path)
			if err != nil {
				return nil, err
			}
		} else {
			fh, err = os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0660)
			if err != nil {
				return nil, err
			}
		}
	}

	f := new(fileLogger)
	f.removeNewlines = removeNewlines
	f.fileHandle = fh
	return f, nil
}

//LaunchModule is intended to run in a separate goroutine and used by rlog internally. It writes log
//messages to file Arguments: [dataChan] Channel to receive log messages. [flushChan] Channel to
//receive flush command
func (conf *fileLogger) LaunchModule(dataChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {

	prefix := common.SyslogHeader()

	//Wait forever on data and flush channel
	for {
		select {
		case logMsg := <-dataChan:
			//Received log message, print it
			conf.writeMsg(logMsg, prefix)
		case ret := <-flushChan:
			//Flush and return success
			conf.flush(dataChan, prefix)
			ret <- true
		}
	}
}

//writeMsg writes message to file
func (conf *fileLogger) writeMsg(rawRlogMsg *common.RlogMsg, prefix string) {
	_, err := conf.fileHandle.WriteString(common.FormatMessage(rawRlogMsg, prefix, conf.removeNewlines) + "\n")
	if !conf.loggedError && err != nil {
		//Log logging error once (only once, otherwise we get a feedback loop)
		conf.loggedError = true
		rlog.Error("[rlog] Unable to write log message to file, reason: " + err.Error())
	}
}

//flush writes all pending log messages to file
//Arguments:[dataChan] data channel to access all pending messages, [prefix] log prefix
func (conf *fileLogger) flush(dataChan <-chan (*common.RlogMsg), prefix string) {
	for {
		//Perform non blocking read until the channel is empty
		select {
		case logMsg := <-dataChan:
			conf.writeMsg(logMsg, prefix)
		default:
			return
		}
	}

	//Do not handle error, as there is nothing we can do about it
	conf.fileHandle.Sync()
}
