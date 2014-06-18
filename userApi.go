package rlog

import (
	"container/list"
	"fmt"
	"github.com/rightscale/rlog/common"
	"log"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

//===== severity levels map to a couple of constants =====
const (
	SeverityFatal   common.RlogSeverity = iota
	SeverityError   common.RlogSeverity = iota
	SeverityWarning common.RlogSeverity = iota
	SeverityInfo    common.RlogSeverity = iota
	SeverityDebug   common.RlogSeverity = iota
)

//===== Data types =====

//logger is an empty struct because the rlog functions on top of it are all
//referring to the singleton rlog instance.
type logger struct{}

//RlogConfig holds the logger configuration. It allows rlog users to configure the logger.
type RlogConfig struct {
	ChanCapacity       uint32 //Buffer capacity for communication between logger and each module
	FlushTimeout       uint32 //Max time for rlog modules to write-back their data (seconds)
	Severity           common.RlogSeverity
	tagsDisabledExcept map[string]bool //All except the listed tags are disabled
	tagsEnabledExcept  map[string]bool //All tags are filtered except for the listed tags
}

//rlogModule interface is implemented by output modules. It requires a function which takes a message
//and a flush channel as argument. When rlog is launched and the module is enabled, this
//function is launched as separate goroutine.
type rlogModule interface {
	LaunchModule(<-chan (*common.RlogMsg), chan (chan (bool)))
}

//===== rlog global data =====

//Keep reference to module initialization functions to launch them as soon as the logger is started
var activeModules *list.List = list.New()

//Initialized stores whether the logger has been initialized
var initialized bool = false

//rlogConfig holds the logger configuration
var config RlogConfig

//A variable for ID generation. Access it ONLY using thread safe methods from sync/atomic!
var uniqueMsgID uint64

//===== Initialization functions =====

//Newlogger creates a new instance of the logger struct. The entire interface for writing
//log messages is available on top of a logger and calls the singleton rlog instance. In contrast
//to using the rlog package directly, a logger can satisfy a log interface required by an
//external library and so decouple the rlog package from the library logger.
func NewLogger() *logger {
	return new(logger)
}

//GetDefaultConfig returns a default configuration for the core logger. Only logging to syslog is activated
//(to be implemented).
//Returns: struct holding default configuration
func GetDefaultConfig() RlogConfig {
	var conf RlogConfig
	conf.ChanCapacity = 100
	conf.FlushTimeout = 2
	conf.Severity = SeverityInfo

	return conf
}

//Start configures the logger and launches it. Once the logger is started, it cannot be started again.
//Start is not thread safe: use Start before spawning any goroutine using the logger.
//Arguments: logger configuration.
func Start(conf RlogConfig) {

	if !initialized {
		//Set configuration and launch modules
		config = conf

		//Initialize the ID generation service to some large number so that it can be found easily
		//in the logs when using grep.
		uniqueMsgID = generateRandomNumber()

		//Now that the configuration is set, we can launch the modules
		launchAllModules()

		initialized = true
	} else {
		Error("Logger initialization triggered but logger already initialized")
	}
}

//EnableModule activates an output module
//Arguments: module to be activated, must implement the rlogModule interface
func EnableModule(module rlogModule) {
	if initialized {
		// Do not allow modification if logger already initialized
		Error("Cannot modify StdoutModuleConfig when logger already running")
	} else {
		//Launch module
		activeModules.PushBack(module)
	}
}

//launchAllModules starts all enabled modules. An enabled module is not launched
//immediately because the arguments to be passed in depend on the rlog core
//configuration. More precisely: the modules require a data and flush channel. The
//channel configuration is set by the user when setting the core configuration. However,
//the core configuration is set when rlog is started which is after enabling the modules.
func launchAllModules() {
	for e := activeModules.Front(); e != nil; e = e.Next() {
		//Cycle over all registered modules and active them
		c, ok := e.Value.(rlogModule)
		if ok {
			go c.LaunchModule(getMsgChannel(), getFlushChannel())
		} else {
			log.Panic("[RightLog4Go FATAL] type assertion for module channel failed\n")
		}
	}
}

//===== Configuration API =====
// converts the given string value to log level (severity).
//
// value: to convert
func (c *RlogConfig) SeverityFromString(value string) {
	switch strings.ToLower(value) {
	case "fatal":
		c.Severity = SeverityFatal
	case "error":
		c.Severity = SeverityError
	case "warning":
		c.Severity = SeverityWarning
	case "info":
		c.Severity = SeverityInfo
	case "debug":
		c.Severity = SeverityDebug
	default:
		panic(fmt.Sprintf("Unknown severity: %s", value))
	}
}

//EnableTagsExcept enables output for all messages except the ones carrying one of the tags
//specified. Using "EnableTagsExcept" overwrites the settings from "DisableTagsExcept".
func (c *RlogConfig) EnableTagsExcept(tags []string) {
	c.tagsDisabledExcept = nil
	c.tagsEnabledExcept = createAndFillStringHt(tags)
}

//DisableTagsExcept enables output for messages carrying one of the tags specified. All other log
//messages are filtered. Using "DisableTagsExcept" overwrites the settings from "EnableTagsExcept".
func (c *RlogConfig) DisableTagsExcept(tags []string) {
	c.tagsDisabledExcept = createAndFillStringHt(tags)
	c.tagsEnabledExcept = nil
}

//createAndFillStringHt creates a hash map and fills it with the elements from the given slice
func createAndFillStringHt(tags []string) map[string]bool {
	ht := make(map[string]bool)
	for _, e := range tags {
		ht[e] = true
	}

	return ht
}

//===== Logging API no tags =====

//Fatal logs a message of severity "fatal".
//Arguments: printf formatted message
func Fatal(format string, a ...interface{}) {
	genericLogHandler("FATAL", "", format, a, SeverityFatal, true)
}

//Fatal logs a message of severity "fatal".
//Arguments: printf formatted message
func (l logger) Fatal(format string, a ...interface{}) {
	genericLogHandler("FATAL", "", format, a, SeverityFatal, true)
}

//Error logs a message of severity "error".
//Arguments: printf formatted message
func Error(format string, a ...interface{}) {
	genericLogHandler("ERROR", "", format, a, SeverityError, true)
}

//Error logs a message of severity "error".
//Arguments: printf formatted message
func (l logger) Error(format string, a ...interface{}) {
	genericLogHandler("ERROR", "", format, a, SeverityError, true)
}

//Warning logs a message of severity "warning".
//Arguments: printf formatted message
func Warning(format string, a ...interface{}) {
	genericLogHandler("WARNING", "", format, a, SeverityWarning, false)
}

//Warning logs a message of severity "warning".
//Arguments: printf formatted message
func (l logger) Warning(format string, a ...interface{}) {
	genericLogHandler("WARNING", "", format, a, SeverityWarning, false)
}

//Info logs a message of severity "info".
//Arguments: printf formatted message
func Info(format string, a ...interface{}) {
	genericLogHandler("INFO", "", format, a, SeverityInfo, false)
}

//Info logs a message of severity "info".
//Arguments: printf formatted message
func (l logger) Info(format string, a ...interface{}) {
	genericLogHandler("INFO", "", format, a, SeverityInfo, false)
}

//Debug logs a message of severity "debug".
//Arguments: printf formatted message
func Debug(format string, a ...interface{}) {
	genericLogHandler("DEBUG", "", format, a, SeverityDebug, false)
}

//Debug logs a message of severity "debug".
//Arguments: printf formatted message
func (l logger) Debug(format string, a ...interface{}) {
	genericLogHandler("DEBUG", "", format, a, SeverityDebug, false)
}

//===== Logging API with tags =====

//FatalT logs a message of severity "fatal".
//Arguments: tag and printf formatted message
func FatalT(tag string, format string, a ...interface{}) {
	genericLogHandler("FATAL", tag, format, a, SeverityFatal, true)
}

//FatalT logs a message of severity "fatal".
//Arguments: tag and printf formatted message
func (l logger) FatalT(tag string, format string, a ...interface{}) {
	genericLogHandler("FATAL", tag, format, a, SeverityFatal, true)
}

//ErrorT logs a message of severity "error".
//Arguments: tag and printf formatted message
func ErrorT(tag string, format string, a ...interface{}) {
	genericLogHandler("ERROR", tag, format, a, SeverityError, true)
}

//ErrorT logs a message of severity "error".
//Arguments: tag and printf formatted message
func (l logger) ErrorT(tag string, format string, a ...interface{}) {
	genericLogHandler("ERROR", tag, format, a, SeverityError, true)
}

//WarningT logs a message of severity "warning".
//Arguments: tag and printf formatted message
func WarningT(tag string, format string, a ...interface{}) {
	genericLogHandler("WARNING", tag, format, a, SeverityWarning, false)
}

//WarningT logs a message of severity "warning".
//Arguments: tag and printf formatted message
func (l logger) WarningT(tag string, format string, a ...interface{}) {
	genericLogHandler("WARNING", tag, format, a, SeverityWarning, false)
}

//InfoT logs a message of severity "info".
//Arguments: tag and printf formatted message
func InfoT(tag string, format string, a ...interface{}) {
	genericLogHandler("INFO", tag, format, a, SeverityInfo, false)
}

//InfoT logs a message of severity "info".
//Arguments: tag and printf formatted message
func (l logger) InfoT(tag string, format string, a ...interface{}) {
	genericLogHandler("INFO", tag, format, a, SeverityInfo, false)
}

//DebugT logs a message of severity "debug".
//Arguments: tag and printf formatted message
func DebugT(tag string, format string, a ...interface{}) {
	genericLogHandler("DEBUG", tag, format, a, SeverityDebug, false)
}

//DebugT logs a message of severity "debug".
//Arguments: tag and printf formatted message
func (l logger) DebugT(tag string, format string, a ...interface{}) {
	genericLogHandler("DEBUG", tag, format, a, SeverityDebug, false)
}

//===== Logging API: tools =====

//GenerateID creates a unique ID, i.e. two calls to GenerateID are guaranteed to return different IDs
//Returns: unique ID
func GenerateID() string {
	id := atomic.AddUint64(&uniqueMsgID, 1)
	return fmt.Sprintf("%x", id)

}

//GenerateID creates a unique ID, i.e. two calls to GenerateID are guaranteed to return different IDs
//Returns: unique ID
func (l logger) GenerateID() string {
	return GenerateID()
}

//Flush should be called before the program using RightLog4Go exits (e.g. by using defer in main).
//Flush notifies the registered logger modules to write back their buffered data.
func Flush() {
	for e := flushChannels.Front(); e != nil; e = e.Next() {
		//Cycle over all registered channels, perform a type conversion because of the linked list
		// and call the helper function implementing the flush protocol
		c, ok := e.Value.(chan chan (bool))
		if ok {
			flushHelper(c)
		} else {
			log.Printf("[RightLog4Go FATAL] type assertion for flush channel failed\n")
		}
	}
}

// Performs a reset of rlog state, intended for testing purposes only (with or
// without flushing before-hand). Applications should call Flush() but should
// usually not reset state. A reset is needed for unit testing due to rlog being
// a singleton. Tests that leverage rlog therefore cannot be run in parallel and
// also call reset state.
func ResetState() {
	if initialized {
		config = *new(RlogConfig)
		msgChannels = list.New()
		flushChannels = list.New()
		activeModules = list.New()
		initialized = false
	}
}

//===== Tools =====

//generateRandomNumber generates a random number
//Returns: random number between 256 and 4194560
func generateRandomNumber() uint64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return uint64((r.Int63n(1<<14) + 1) << 8)
}
