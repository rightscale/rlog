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
	"strings"
)

//Configuration of syslog module
type syslogModuleConfig struct {
	syslogConn *goSyslog.Writer
}

//Define constant for logging to syslog on localhost or remote logging
//Not yet exposed
const (
	syslogLocalhost string = ""
	syslogUnix      string = ""
	syslogTCP       string = "tcp"
	syslogUDP       string = "udp"
)

//NewSyslogLogger enables logging to syslog.
//Returns: instance of syslog logger module in case of success, error otherwise
func NewLocalSyslogLogger() (*syslogModuleConfig, error) {

	//Exposing these parameters to the user is currently not implemented.
	//Set it to the Go defaults for "log to local syslog server"
	//Arguments (not yet implemented): [network] connection type, can be: SyslogTCP and SyslogUDP, [addr] target
	//syslog server addr. Use SyslogLocalhost constant to log to syslog on local host
	var network string = syslogUnix
	var addr string = syslogLocalhost

	var err error
	conf := new(syslogModuleConfig)

	conf.syslogConn, err = goSyslog.Dial(network, addr, goSyslog.LOG_INFO, path.Base(os.Args[0]))
	if err != nil {
		log.Printf("Could not open connection to syslog, reason: " + err.Error())
		return nil, err
	} else if conf.syslogConn == nil {
		log.Printf("Could not retrieve connection to syslog")
		return nil, fmt.Errorf("Could not retrieve connection to syslog")
	} else {
		conf.syslogConn.Debug("rlog syslog module started successfully")
		return conf, nil
	}
}

//LaunchModule is intended to run in a separate goroutine. It prints log messages to syslog
//Arguments: [dataChan] Channel to receive log messages. [flushChan] Channel to receive flush command
func (conf *syslogModuleConfig) LaunchModule(dataChan <-chan (*common.RlogMsg), flushChan chan (chan (bool))) {

	//Wait forever on data and flush channel
	for {
		select {
		case logMsg := <-dataChan:
			//Received log message, print it
			conf.syslogProcessMessage(logMsg)
		case ret := <-flushChan:
			//Flush and return success
			conf.syslogFlush(dataChan)
			ret <- true
		}
	}
}

//syslogProcessMessage prints the message to syslog
//Arguments: log message
func (conf *syslogModuleConfig) syslogProcessMessage(m *common.RlogMsg) {

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

	//Write log message using appropriate syslog severity level
	switch m.Severity {
	case rlog.SeverityDebug:
		conf.syslogConn.Debug(logMsg)
	case rlog.SeverityInfo:
		conf.syslogConn.Info(logMsg)
	case rlog.SeverityWarning:
		conf.syslogConn.Warning(logMsg)
	case rlog.SeverityError:
		conf.syslogConn.Err(logMsg)
	case rlog.SeverityFatal:
		conf.syslogConn.Crit(logMsg)
	}
}

//syslogFlush writes all pending log messages to syslog
//Arguments: data channel to access all pending messages
func (conf *syslogModuleConfig) syslogFlush(dataChan <-chan (*common.RlogMsg)) {
	for {
		//Read from data channel until there is nothing more to read, then return
		select {
		case logMsg := <-dataChan:
			conf.syslogProcessMessage(logMsg)
		default:
			return
		}
	}
}
