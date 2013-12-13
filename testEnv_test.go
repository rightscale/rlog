/*
Helper functions for testing
*/
package rlog

import (
	"container/list"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"log"
	"runtime"
	"testing"
)

//Hook this testing framework into go test
func Test(t *testing.T) { TestingT(t) }

//===========  Test suite with uninitialized logger =====

type Uninitialized struct{}

var _ = Suite(&Uninitialized{})

//before test: disable output from internal go logger
func (s *Uninitialized) SetUpSuite(c *C) {
	disableGoLog()

	//Allocate multiple system threads to allow for real concurrency
	runtime.GOMAXPROCS(10)
}

//before each: logger is reset and is uninitialized
func (s *Uninitialized) SetUpTest(c *C) {
	resetState()
}

//===== Test suite with initialized logger =====

type Initialized struct{}

var _ = Suite(&Initialized{})

//before test: disable output from internal go logger
func (s *Initialized) SetUpSuite(c *C) {
	disableGoLog()

	//Allocate multiple system threads to allow for real concurrency
	runtime.GOMAXPROCS(10)
}

//before each: logger is reset and is initialized
func (s *Initialized) SetUpTest(c *C) {
	resetAndInitialize()
}

//===== Setup test suite to test stateless functions with no overhead =====

type Stateless struct{}

var _ = Suite(&Stateless{})

//before test: disable output from internal go logger
//before test: logger is reset and is initialized
func (s *Stateless) SetUpSuite(c *C) {
	disableGoLog()
	resetState()

	//Allocate multiple system threads to allow for real concurrency
	runtime.GOMAXPROCS(10)
}

//===== Helper functions =====

//resetState resets package global variables to their initial state for testing purpose
func resetState() {
	initialized = false
	config = *new(RlogConfig)
	msgChannels = list.New()
	flushChannels = list.New()
	activeModules = list.New()
}

//resetAndInitialize resets the logger to its default configuration and initializes it
func resetAndInitialize() {
	resetState()
	conf := GetDefaultConfig()
	conf.Severity = SeverityDebug
	Start(conf)
}

//disableGoLog redirects all output from the go logging package to dev/null
func disableGoLog() {
	log.SetOutput(ioutil.Discard)
}
