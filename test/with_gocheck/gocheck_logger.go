/*
Implements an on-failure-only gocheck package logger.
*/
package with_gocheck

import (
	"github.com/rightscale/rlog"
	"github.com/rightscale/rlog/common"
	"launchpad.net/gocheck"
)

// Test logger that works for any test harness built on top of testing package.
type GoCheckLogger struct {
	c *gocheck.C
}

// Creates a logger using gocheck object.
//
// t: testing object
//
// return: instance of test logger
func NewGoCheckLogger(c *gocheck.C) *GoCheckLogger {
	return &GoCheckLogger{c}
}

// Convenience method to initialize rlog with a single (error-level) gocheck
// logger and start rlog. Can be called at the start of your test method or in
// your test setup. Remember to put "defer rlog.Flush()" either in your test
// method(s) or test teardown method. The test teardown is invoked before the
// success/failure of the gocheck test is evaluated.
func StartGoCheckLogger(c *gocheck.C) {
	rlog.ResetState()
	rlog.EnableModule(NewGoCheckLogger(c))
	rlogConf := rlog.GetDefaultConfig()
	rlogConf.Severity = rlog.SeverityError
	rlog.Start(rlogConf)
}

// Intended to run in a separate goroutine. It prints log messages to console.
//
// dataChan: receives log messages.
//
// flushChan: receives flush command.
func (self *GoCheckLogger) LaunchModule(dataChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {

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
func (self *GoCheckLogger) printMsg(rawRlogMsg *common.RlogMsg, prefix string) {
	msg := common.FormatMessage(rawRlogMsg, prefix, false)
	self.c.Log(msg)
}

// Flushes pending messages to console.
//
// dataChan: data channel to access all pending messages
//
// prefix: log prefix
func (self *GoCheckLogger) flush(dataChan <-chan (*common.RlogMsg), prefix string) {
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
