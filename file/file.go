/*
Package file implements an output module for logging to a file using rlog.
*/
package file

import (
	"fmt"
	"github.com/rightscale/rlog/common"
	"os"
	"path/filepath"
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
	f := new(fileLogger)
	f.removeNewlines = removeNewlines
	err := f.openFile(path, overwrite)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// opens the log file using the given criteria.
func (conf *fileLogger) openFile(path string, overwrite bool) error {
	var err error

	parentDir, _ := filepath.Split(path)
	if parentDir != "" {
		var dirMode os.FileMode = 0770 // user/group-only read/write/traverse
		err = os.MkdirAll(parentDir, dirMode)
		if err != nil {
			return err
		}
	}

	// open write-only (will never read back from log file).
	var fh *os.File
	var fileMode os.FileMode = 0660 // user/group-only read/write

	if overwrite {
		// create or truncate
		// note that os.Create() is too permissive (i.e. grants world read/write).
		fh, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return err
		}
	} else {
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			// not present, create it
			fh, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, fileMode)
			if err != nil {
				return err
			}
		} else {
			// append to existing
			fh, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, fileMode)
			if err != nil {
				return err
			}
		}
	}
	conf.fileHandle = fh
	return nil
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
			err := conf.writeMsg(logMsg, prefix)
			if err != nil {
				// we may be able to work around intermittent failures by reopening.
				if conf.reopenFile() != nil {
					err = conf.writeMsg(logMsg, prefix)
				}
			}
			if err != nil {
				// panic if reopening did not resolve the issue.
				panic(err)
			}
		case ret := <-flushChan:
			//Flush and return success
			conf.flush(dataChan, prefix)
			ret <- true
		}
	}
}

//writeMsg writes message to file
func (conf *fileLogger) writeMsg(rawRlogMsg *common.RlogMsg, prefix string) error {
	_, err := fmt.Fprintln(conf.fileHandle, common.FormatMessage(rawRlogMsg, prefix, conf.removeNewlines))
	return err
}

//flush writes all pending log messages to file
//Arguments:[dataChan] data channel to access all pending messages, [prefix] log prefix
func (conf *fileLogger) flush(dataChan <-chan (*common.RlogMsg), prefix string) {

	// we may already be panicking due to losing file handle.
	if conf.fileHandle == nil {
		return
	}

	// reopen file before flushing any messages to support rotation of file logs
	// in response to SIGHUP, etc.
	err := conf.reopenFile()
	if err != nil {
		// panic if unable to reopen log file so that service can be restarted by
		// outer harness with alerts, etc.
		panic(err)
	}

	for {
		//Perform non blocking read until the channel is empty
		select {
		case logMsg := <-dataChan:
			err = conf.writeMsg(logMsg, prefix)
			if err != nil {
				// we reopened before we began flushing so any failure during flush
				// cannot logically be resolved by reopening again here.
				panic(err)
			}
		default:
			return
		}
	}

	//Do not handle error, as there is nothing we can do about it
	conf.fileHandle.Sync()
}

// reopen existing log file and/or create new file if log rotation renamed
// existing file.
func (conf *fileLogger) reopenFile() error {
	// note that the trick here is that the file struct remembers the original
	// file name before it was renamed by rotation, if ever.
	oldFileHandle := conf.fileHandle
	conf.fileHandle = nil
	path := oldFileHandle.Name()
	err := oldFileHandle.Close()
	if err == nil {
		err = conf.openFile(path, false)
	}

	return err
}
