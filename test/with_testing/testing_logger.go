/*
Implements an on-failure-only testing package logger.
*/
package with_testing

import (
	"github.com/rightscale/rlog"
	"github.com/rightscale/rlog/common"
	"testing"
)

// Test logger that works for any test harness built on top of testing package.
type TestingLogger struct {
	t *testing.T
}

// Creates a logger using testing object.
//
// t: testing object
//
// return: instance of test logger
func NewTestingLogger(t *testing.T) *TestingLogger {
	return &TestingLogger{t}
}

// Convenience method to initialize rlog with a single (error-level) testing
// logger and start rlog. Remember to put "defer rlog.Flush()" somewhere in your
// test method(s).
func StartTestingLogger(t *testing.T) {
	rlog.ResetState()
	rlog.EnableModule(NewTestingLogger(t))
	rlogConf := rlog.GetDefaultConfig()
	rlogConf.Severity = rlog.SeverityError
	rlog.Start(rlogConf)
}

// Intended to run in a separate goroutine. It prints log messages to console.
//
// dataChan: receives log messages.
//
// flushChan: receives flush command.
func (self *TestingLogger) LaunchModule(dataChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {

	prefix := common.SyslogHeader()

	// wait forever on data and flush channel
	for {
		select {
		case logMsg := <-dataChan:
			// received log message, print it
			self.printMsg(logMsg, prefix)
		case ret := <-flushChan:
			// flush and return success
			self.flush(dataChan, prefix)
			ret <- true
		}
	}
}

// Prints the message to console.
//
// rawRlogMsg: log message received from channel.
//
// prefix: log prefix
func (self *TestingLogger) printMsg(rawRlogMsg *common.RlogMsg, prefix string) {
	msg := common.FormatMessage(rawRlogMsg, prefix, false)
	// note that t.Log() entry is unconditionally prefixed with this file and line
	// number, so embed a newline to make it easier to distinguish message.
	self.t.Logf("\n%s", msg)
}

// Flushes pending messages to console.
//
// dataChan: data channel to access all pending messages
//
// prefix: log prefix
func (self *TestingLogger) flush(dataChan <-chan (*common.RlogMsg), prefix string) {
	for {
		// perform non blocking read until the channel is empty
		select {
		case logMsg := <-dataChan:
			self.printMsg(logMsg, prefix)
		default:
			return
		}
	}
}
