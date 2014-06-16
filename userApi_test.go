/*
These tests cover:
- Initialization
- Integration testing: use logging functions and intercept channel communication
  ==> "guarantees" that message actually appears at the other side
*/
package rlog

import (
	"container/list"
	"github.com/rightscale/rlog/common"
	. "launchpad.net/gocheck"
	"strings"
)

type fakeLogModule struct {
	msgChan   <-chan (*common.RlogMsg)
	flushChan chan (chan (bool))
}

func (f *fakeLogModule) LaunchModule(msgChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {
	f.msgChan = msgChan
	f.flushChan = flushChan
}

//Test initialization
func (s *Uninitialized) TestStart(t *C) {

	EnableModule(new(fakeLogModule))
	conf := GetDefaultConfig()
	conf.ChanCapacity = 101

	//When calling start, it should (1) set the logger state to initialized
	Start(conf)
	if initialized == false {
		t.Fatalf("Initialization variable not set")
	}

	//(2) apply the given configuration
	if config.ChanCapacity != 101 {
		t.Fatalf("Initialize did not apply capacity configuration")
	}

	//(3) Create at least one communication channel because stdout logging is enabled
	if msgChannels.Front() == nil {
		t.Fatalf("Channel not initialized after initialization")
	}

	//(4) uniqueMsgID gets set to a number with at least three digits
	if uniqueMsgID < 100 {
		t.Fatalf("uniqueMsgID start value is < 100 but it should initialized larger during Initialization")
	}

	//Hook in our own channel to intercept messages for testing
	msgChannels = list.New()
	c := getMsgChannel()

	//When the logger is initialized a second time, it should generate an error log entry
	Start(GetDefaultConfig())
	logMsg := nonBlockingChanRead(c) //Check channel for error msg
	if logMsg == nil {
		t.Fatalf("Error message not generated after double initialization")
	} else if !strings.Contains(logMsg.Msg, "logger already initialized") {
		t.Fatalf("Initializing twice did not generate an error")
	}
}

//When generating two IDs, it should create different ones
func (s *Stateless) TestIDGeneration(t *C) {

	if GenerateID() == GenerateID() {
		t.Fatalf("Two ID generation function calls resulted in the same ID, but they should be different")
	}
}

//Test the various logging routines. This is for integration testing, as the various sub components like
//channels, msg formatting are tested independently.
func (s *Initialized) TestLoggingRoutines(t *C) {

	//Create message for comparison
	msg := "testmessage 10"

	//Create our own destination channel for testing purpose
	msgChannels = list.New()
	myChan := getMsgChannel()

	//When printing an Error message, it should generate an Error message and push it to the channel
	Fatal("testmessage %d", 10)
	logFunctionVerify(t, SeverityFatal, false, msg, myChan)

	//When printing an Error message, it should generate an Error message and push it to the channel
	Error("testmessage %d", 10)
	logFunctionVerify(t, SeverityError, false, msg, myChan)

	//When printing an Info message, it should generate an Info message and push it to the channel
	Info("testmessage %d", 10)
	logFunctionVerify(t, SeverityInfo, false, msg, myChan)

	//When printing an Error message, it should generate an Error message and push it to the channel
	Debug("testmessage %d", 10)
	logFunctionVerify(t, SeverityDebug, false, msg, myChan)
}

//Test the various logging routines defined on top of log objects.
func (s *Initialized) TestLogObjectRoutines(t *C) {

	//Create message for comparison
	msg := "logger object test message 20"

	//Create our own destination channel for testing purpose
	msgChannels = list.New()
	myChan := getMsgChannel()

	//Create a log object
	myLogger := NewLogger()

	//When printing an Error message, it should generate an Error message and push it to the channel
	myLogger.Fatal("logger object test message %d", 20)
	logFunctionVerify(t, SeverityFatal, false, msg, myChan)

	//When printing an Error message, it should generate an Error message and push it to the channel
	myLogger.Error("logger object test message %d", 20)
	logFunctionVerify(t, SeverityError, false, msg, myChan)

	//When printing an Info message, it should generate an Info message and push it to the channel
	myLogger.Info("logger object test message %d", 20)
	logFunctionVerify(t, SeverityInfo, false, msg, myChan)

	//When printing an Error message, it should generate an Error message and push it to the channel
	myLogger.Debug("logger object test message %d", 20)
	logFunctionVerify(t, SeverityDebug, false, msg, myChan)

	//Test ID generation service
	id1 := myLogger.GenerateID()
	id2 := myLogger.GenerateID()
	if id1 == id2 {
		t.Fatalf("It should generate two different ids, got: id1=%s and id2=%s", id1, id2)
	}
}

//logFunctionVerify is a generic function which fetches a log message directly from the channel (if
//a log msg is there) and matches it against the expectation of the log printing function (info, error, etc.)
//called before.
//Parameter: [t] testing framework param. [severity] Expected severity. [msgPresent] expect message to be present
//(may be filtered)? [msg] Expected message. [myChan] Mocked log destination.
func logFunctionVerify(t *C, severity common.RlogSeverity, msgPresent bool, msg string, myChan <-chan (*common.RlogMsg)) {

	rlm := nonBlockingChanRead(myChan)
	if rlm == nil {
		if !msgPresent {
			//When a message is filtered, we do not expect a message to be logged. However, in this case
			//a message should have been present according to the msgPresent flag.
			t.Fatalf("Expected log message but did not receive a message")
		}
	} else {
		if msgPresent {
			t.Fatalf("This message should have been filtered according to its tag, msg: %s", rlm.Msg)
		}
		if !strings.Contains(rlm.Msg, msg) {
			t.Fatalf("Log message does not contain actual message, msg: %s, log: %s", msg, rlm.Msg)
		}
		if rlm.Severity != severity {
			t.Fatalf("Severity parameter should be %d but is %d", severity, rlm.Severity)
		}
		if severity == SeverityFatal || severity == SeverityError {
			if rlm.StackTrace == "" {
				t.Fatalf("Stack trace should be present, but it is not")
			}
		}
	}
}
