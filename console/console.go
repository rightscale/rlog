/*
Implements a console logger for stdout/stderr.
*/
package console

import (
  "fmt"
  "github.com/brsc/rlog/common"
  "os"
)

// Console logger (type exported for deprecated stdout module but fields are private).
type ConsoleLogger struct {
  removeNewlines bool
  outputFile     *os.File
}

// Creates a logger for stdout.
//
// removeNewlines: true to replace newlines
//
// return: instace of console logger
func NewStdoutLogger(removeNewlines bool) *ConsoleLogger {
  logger := new(ConsoleLogger)
  logger.removeNewlines = removeNewlines
  logger.outputFile = os.Stdout
  return logger
}

// Creates a logger for stderr.
//
// removeNewlines: true to replace newlines
//
// return: instace of console logger
func NewStderrLogger(removeNewlines bool) *ConsoleLogger {
  logger := new(ConsoleLogger)
  logger.removeNewlines = removeNewlines
  logger.outputFile = os.Stderr
  return logger
}

// Intended to run in a separate goroutine. It prints log messages to console.
//
// dataChan: receives log messages.
//
// flushChan: receives flush command.
func (conf *ConsoleLogger) LaunchModule(dataChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {

  prefix := common.SyslogHeader()

  // wait forever on data and flush channel
  for {
    select {
    case logMsg := <-dataChan:
      // received log message, print it
      conf.printMsg(logMsg, prefix)
    case ret := <-flushChan:
      // flush and return success
      conf.flush(dataChan, prefix)
      ret <- true
    }
  }
}

// Prints the message to console.
//
// rawRlogMsg: log message received from channel.
//
// prefix: log prefix
func (conf *ConsoleLogger) printMsg(rawRlogMsg *common.RlogMsg, prefix string) {
  msg := common.FormatMessage(rawRlogMsg, prefix, conf.removeNewlines)
  fmt.Fprintln(conf.outputFile, msg)
}

// Flushes pending messages to console.
//
// dataChan: data channel to access all pending messages
//
// prefix: log prefix
func (conf *ConsoleLogger) flush(dataChan <-chan (*common.RlogMsg), prefix string) {
  for {
    // perform non blocking read until the channel is empty
    select {
    case logMsg := <-dataChan:
      conf.printMsg(logMsg, prefix)
    default:
      return
    }
  }
}
