/*
These tests cover:
- Message formatting
- Stack trace creation
- File and position calculation
*/
package rlog

import (
  "github.com/brsc/rlog/common"
  . "launchpad.net/gocheck"
  "runtime"
  "strconv"
  "strings"
)

//Test log header formatting
func (s *Stateless) TestFormatHeaders(t *C) {
  level := "testLevel"
  tag := "foo.bar"
  file := "test/testfile.go"
  line := 10

  //When posInfo set to true, level, file and line should appear in the log header
  header := formatHeaders(true, level, tag, file, line)
  if !strings.Contains(header, level) {
    t.Fatalf("Expected log level in header. but header is only: " + header)
  }
  if !strings.Contains(header, tag) {
    t.Fatalf("Expected tag in header. but header is only: " + header)
  }
  if !strings.Contains(header, file) {
    t.Fatalf("Expected file name in header. but header is only: " + header)
  }
  if !strings.Contains(header, strconv.Itoa(line)) {
    t.Fatalf("Expected line number in header. but header is only: " + header)
  }

  //When posInfo set to false, level should appear in log header but
  //file and line should not appear in log header
  header = formatHeaders(false, level, tag, file, line)
  if !strings.Contains(header, level) {
    t.Fatalf("Expected log level in header. but header is only: " + header)
  }
  if !strings.Contains(header, tag) {
    t.Fatalf("Expected tag in header. but header is only: " + header)
  }
  if strings.Contains(header, file) {
    t.Fatalf("Expected no file name in header. but header is only: " + header)
  }
  if strings.Contains(header, strconv.Itoa(line)) {
    t.Fatalf("Expected no line number in header. but header is only: " + header)
  }
}

//When generateLogMessage is invoked, it should create a log message with the appropriate flags set
func (s *Stateless) TestGenerateLogMessage(t *C) {
  generateLogMessage_helper(t, SeverityError)
  generateLogMessage_helper(t, SeverityInfo)
}

//generateLogMessage_helper tests the generateLogMsg algorithm.
//Parameters: [t] Testing framework. [severity] Expected severity level
func generateLogMessage_helper(t *C, severity common.RlogSeverity) {
  level := "testLevel"
  tag := "hmmm"
  msg := "testMessage"
  file := "test/testfile.go"
  line := 10
  pc := uint(200)

  rawTestInfo := logPieces{level, tag, msg, severity, false, file, line, pc, "trace"}
  rlm := rawTestInfo.generateLogMsg()
  if rlm.Pc != pc {
    t.Fatalf("Expected PC to be %d, but it is: %d", pc, rlm.Pc)
  }
  if rlm.Severity != severity {
    t.Fatalf("Expected severity to be %d, but it is: %d", severity, rlm.Severity)
  }
  if !strings.Contains(rlm.Msg, msg) {
    t.Fatalf("Expected message to contain \"%s\", but message is: %s", msg, rlm.Msg)
  }
  if !strings.Contains(rlm.StackTrace, "trace") {
    t.Fatalf("Log message struct does not hold stack trace")
  }
}

//When the logger is not initialized, writing log messages should fail
func (*Uninitialized) TestGenericLogHandler(t *C) {
  level := "testLevel"
  tag1 := "testTag1"

  format, params := simulatePrintf("test - %d\n", 10)
  ret := genericLogHandler(level, tag1, format, params, SeverityError, false)
  if ret {
    t.Fatalf("genericLogHandler should have failed because the logger was not initialized")
  }
}

//When creating a log entry, it should fetch the correct file and line number
func (s *Stateless) TestGetLogCallPos(t *C) {

  file, line, logMsg := getCurrentStackEnvironment()

  if !strings.Contains(logMsg.Msg, file) {
    t.Fatalf("Error log message does not contain correct file path (or no file path). Expecting: %s, msg: %s", file, logMsg.Msg)
  }

  if !strings.Contains(logMsg.Msg, line) {
    t.Fatalf("Error log message does not contain correct line in file (or no line). Expecting %s, msg: %s", line, logMsg.Msg)
  }
}

//When creating a log entry accompanied by a stack trace, it should create a stack trace starting at the position
//where the log message was created
func (s *Stateless) TestGetStackTrace(t *C) {

  file, line, logMsg := getCurrentStackEnvironment()

  //Extract first line and compare it to the manually fetched runtime information
  //When splitting after newlines, we skip the first line as that line contains the function call information
  //and not the line number we are looking for to compare against the expected result
  firstLine := strings.SplitAfterN(logMsg.StackTrace, "\n", 5)[1]
  if !strings.Contains(firstLine, file) {
    t.Fatalf("Stack trace does not start with correct file, expected: %s, got: %s", file, firstLine)
  }
  if !strings.Contains(firstLine, line) {
    t.Fatalf("Stack trace does not have correct line number, expected: %s, got: %s", line, firstLine)
  }
}

func (s *Initialized) TestIsFilteredSeverity(t *C) {
  config.SeverityFromString("warn")

  //It should filter debug and info
  t.Assert(isFilteredSeverity(SeverityDebug), Equals, true)
  t.Assert(isFilteredSeverity(SeverityInfo), Equals, true)
  t.Assert(isFilteredSeverity(SeverityWarn), Equals, false)
  t.Assert(isFilteredSeverity(SeverityError), Equals, false)
  t.Assert(isFilteredSeverity(SeverityFatal), Equals, false)
}

func (s *Initialized) TestIsFilteredTag(t *C) {
  const tag1 string = "tag1"
  const tag2 string = "tag2"

  //Test EnableTagsExcept
  config.EnableTagsExcept([]string{tag1})
  t.Assert(isFilteredTag(tag1), Equals, true)
  t.Assert(isFilteredTag(tag2), Equals, false)

  //Test DisableTagsExcept
  config.DisableTagsExcept([]string{tag1})
  t.Assert(isFilteredTag(tag1), Equals, false)
  t.Assert(isFilteredTag(tag2), Equals, true)
}

//getCurrentStackEnvironment resets the logger, generates and error message and intercepts it. It furthermore
//fetches the file and line we expect to be present in the log.
//Returns: Expected file and line number to be present in log and the intercepted log message.
func getCurrentStackEnvironment() (string, string, *common.RlogMsg) {
  //Reset state and capture output using our own channel
  resetAndInitialize()
  myChan := getMsgChannel()

  //Obtain information about our file and position (baseline). Afterwards, write error message and intercept it
  _, file, myLine, _ := runtime.Caller(0)
  Error("posTest")
  logMsg := nonBlockingChanRead(myChan)

  //Increment line as Error was called one line after our call to runtime
  myLine++

  return file, strconv.Itoa(myLine), logMsg
}

//simulatePrintf is a helper function converting variadic to slice
func simulatePrintf(format string, a ...interface{}) (string, []interface{}) {
  return format, a
}
