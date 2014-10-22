/*
Package syslog implements an output module for logging to syslog using rlog.
*/
package syslog

import (
	"fmt"
	"github.com/rightscale/rlog"
	"github.com/rightscale/rlog/common"
	"log"
	goSyslog "log/syslog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//Configuration of syslog module
type syslogModuleConfig struct {
	network           string           // one of ["", syslogTCP, syslogUDP]
	raddr             string           // remote syslog server or empty for local
	facility          int              // facility (e.g. LOG_LOCAL0)
	tag               string           // tag for messages or empty for full binary path
	syslogConn        *goSyslog.Writer // writer
	heartBeatFilePath string           // FIX: remove this when we figure out issue with silent syslogger
}

//Define constant for logging to syslog on localhost or remote logging
//Not yet exposed
const (
	maxMessageLength int    = 6 * 1024 // FIX: limited to 6 KB to see if this keeps syslogger humming
	syslogLocalhost  string = ""
	syslogUnix       string = ""
	syslogTCP        string = "tcp"
	syslogUDP        string = "udp"
)

var facilityNames []string = []string{
	"kern", "user", "mail", "daemon", "auth", "syslog", "lpr", "news",
	"uucp", "cron", "security", "ftp", "ntp", "logaudit", "logalert", "clock",
	"local0", "local1", "local2", "local3", "local4", "local5", "local6", "local7"}

//NewLocalSyslogLogger enables logging to syslog.
//Returns: instance of syslog logger module in case of success, error otherwise
func NewLocalSyslogLogger() (*syslogModuleConfig, error) {

	conf := new(syslogModuleConfig)
	err := conf.connectToSyslog(
		syslogUnix,
		syslogLocalhost,
		0, // =LOG_KERN, see NewLocalFacilitySyslogLogger() to select a facility
		path.Base(os.Args[0]))
	if err != nil {
		return nil, err
	}
	return conf, nil
}

//NewSyslogLogger enables logging to syslog with full syslog parameters.
//Params: see syslog.Dial() remarks
//Returns: instance of syslog logger module in case of success, error otherwise
func NewLocalFacilitySyslogLogger(
	network, raddr string,
	facility int,
	heartBeatFilePath string) (*syslogModuleConfig, error) {

	conf := new(syslogModuleConfig)
	conf.heartBeatFilePath = heartBeatFilePath // FIX: strictly for debugging
	err := conf.connectToSyslog(
		network,
		raddr,
		facility,
		path.Base(os.Args[0]))
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// converts given (lowercase) facility name to its integer value equivalent.
func FacilityNameToValue(name string) (int, error) {
	// note that golang as no built-in way to get index from array.
	for idx, n := range facilityNames {
		if n == name {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("Unknown syslog facility name: %s\nMust be one of %v", name, facilityNames)
}

// converts given facility integer value to its (lowercase) name equivalent.
func FacilityValueToName(value int) (string, error) {
	if value < 0 || value >= len(facilityNames) {
		return "", fmt.Errorf(
			"facility is out of range = %d (must be 0-%d).",
			value,
			len(facilityNames)-1)
	}
	return facilityNames[value], nil
}

// establishes the connection to syslog.
func (conf *syslogModuleConfig) connectToSyslog(
	network,
	raddr string,
	facility int,
	tag string) error {

	facilityName, err := FacilityValueToName(facility)
	if err != nil {
		return err
	}

	var priority goSyslog.Priority = goSyslog.Priority(facility<<3) | goSyslog.LOG_INFO

	conf.network = network
	conf.raddr = raddr
	conf.facility = facility
	conf.tag = tag
	conf.syslogConn, err = goSyslog.Dial(network, raddr, priority, tag)

	if err != nil {
		log.Printf("Could not open connection to syslog, reason: " + err.Error())
		return err
	}
	if conf.syslogConn == nil {
		log.Printf("Could not retrieve connection to syslog")
		return fmt.Errorf("Could not retrieve connection to syslog")
	}

	conf.syslogConn.Debug(
		fmt.Sprintf(
			"rlog syslog (re)connected with facility=%d(%s), tag=\"%s\"",
			facility,
			facilityName,
			tag))
	conf.syslogConn.Debug(
		fmt.Sprintf(
			"rlog syslog network=\"%s\", raddr=\"%s\", heartBeatFilePath=\"%s\"",
			network,
			raddr,
			conf.heartBeatFilePath))

	// FIX: heartbeat for debugging only.
	if conf.heartBeatFilePath != "" {
		parentDir, _ := filepath.Split(conf.heartBeatFilePath)
		if parentDir != "" {
			var dirMode os.FileMode = 0775 // user/group-only read/write/traverse, world read/traverse
			err = os.MkdirAll(parentDir, dirMode)
			if err != nil {
				return err
			}
		}
		err = conf.writeHeartBeat("Starting heartbeat...")
		if err != nil {
			return err
		}
	}
	return nil
}

//LaunchModule is intended to run in a separate goroutine. It prints log messages to syslog
//Arguments: [dataChan] Channel to receive log messages. [flushChan] Channel to receive flush command
func (conf *syslogModuleConfig) LaunchModule(dataChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {

	//Wait forever on data and flush channel
	for {
		select {
		case logMsg := <-dataChan:
			//Received log message, print it
			err := conf.syslogProcessMessage(logMsg)
			if err != nil {
				// we may be able to work around intermittent failures by reconnecting.
				if conf.syslogReconnect() != nil {
					err = conf.syslogProcessMessage(logMsg)
				}
			}
			if err != nil {
				// panic if reconnecting did not resolve the issue.
				panic(err)
			}
		case ret := <-flushChan:
			//Flush and return success
			conf.syslogFlush(dataChan)
			ret <- true
		}
	}
}

//syslogProcessMessage prints the message to syslog
//Arguments: log message
func (conf *syslogModuleConfig) syslogProcessMessage(m *common.RlogMsg) error {

	//Prepare log message. Add stack trace of severity is error or fatal
	logMsg := m.Msg
	if m.Severity == rlog.SeverityError || m.Severity == rlog.SeverityFatal {
		logMsg += " -- " + m.StackTrace
	}

	// remove tabs, carriage returns and newlines from any messages sent to syslog
	// due to problems with recording whitespace.
	logMsg = strings.Replace(logMsg, "\t", "", -1)
	logMsg = strings.Replace(logMsg, "\r", "", -1)
	logMsg = strings.Replace(logMsg, "\n", " -- ", -1)

	// FIX: truncate message in attempt to resolve issue with syslog going quiet.
	// not sure what the max datagram size is or if this will help anything...
	if len(logMsg) > maxMessageLength {
		runes := []rune(logMsg)
		logMsg = string(runes[0:maxMessageLength])
	}

	// FIX: write to heartbeat file to determine if this go routine is still
	// running or has been blocked or died silently, etc.
	var err error
	if conf.heartBeatFilePath != "" {
		err = conf.writeHeartBeat(logMsg)
		if err != nil {
			return err
		}
	}

	//Write log message using appropriate syslog severity level
	switch m.Severity {
	case rlog.SeverityDebug:
		err = conf.syslogConn.Debug(logMsg)
	case rlog.SeverityInfo:
		err = conf.syslogConn.Info(logMsg)
	case rlog.SeverityWarning:
		err = conf.syslogConn.Warning(logMsg)
	case rlog.SeverityError:
		err = conf.syslogConn.Err(logMsg)
	case rlog.SeverityFatal:
		err = conf.syslogConn.Crit(logMsg)
	}
	return err
}

//syslogFlush writes all pending log messages to syslog
//Arguments: data channel to access all pending messages
func (conf *syslogModuleConfig) syslogFlush(dataChan <-chan (*common.RlogMsg)) {

	// we may already be panicking due to losing syslog connection.
	if conf.syslogConn == nil {
		return
	}

	// always reestablish syslog connection before flushing message channel to
	// ensure connection liveness (after a day of being open, etc.).
	err := conf.syslogReconnect()
	if err != nil {
		// panic if unable to reestablish connection (rsyslog service is down, etc.)
		// this is useful for a service because it can be restarted by its outer
		// harness with appropriate alerts, etc.
		panic(err)
	}

	for {
		//Read from data channel until there is nothing more to read, then return
		select {
		case logMsg := <-dataChan:
			err = conf.syslogProcessMessage(logMsg)
			if err != nil {
				// we reconnected before we began flushing so any failure during flush
				// cannot logically be resolved by reconnecting again here.
				panic(err)
			}
		default:
			return
		}
	}
}

// closes existing connection and attempts to reconnect to syslog.
func (conf *syslogModuleConfig) syslogReconnect() error {
	oldSyslogConn := conf.syslogConn
	conf.syslogConn = nil
	err := oldSyslogConn.Close()
	if err == nil {
		err = conf.connectToSyslog(conf.network, conf.raddr, conf.facility, conf.tag)
	}

	return err
}

// closes existing connection and attempts to reconnect to syslog.
func (conf *syslogModuleConfig) writeHeartBeat(logMsg string) error {
	var fh *os.File
	var fileMode os.FileMode = 0664 // user/group-only read/write, world read
	var err error

	// always overwrite.
	fh, err = os.OpenFile(conf.heartBeatFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fmt.Fprintln(fh, logMsg)

	return err
}
