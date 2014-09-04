/*
Application modules are test permutations testing various rlog output modules.
*/
package main

import (
	"github.com/rightscale/rlog"
	"github.com/rightscale/rlog/console"
	"github.com/rightscale/rlog/file"
	"github.com/rightscale/rlog/syslog"
	"os"
	"strings"
)

func main() {

	//Setup syslog module
	syslogModule, err := syslog.NewLocalSyslogLogger()
	if err != nil {
		panic("Getting syslog logger instance failed")
	}

	//Setup file logger
	log_file_name := "tmp/test.txt"
	rotated_log_name := log_file_name + ".1"
	if _, err = os.Stat(rotated_log_name); err == nil {
		os.Remove(rotated_log_name)
	}
	fileModule, err := file.NewFileLogger(log_file_name, true, true)
	if err != nil {
		panic("Getting file logger instance failed: " + err.Error())
	}

	rlog.EnableModule(syslogModule)
	rlog.EnableModule(fileModule)
	rlog.EnableModule(console.NewStdoutLogger(true))
	rlog.EnableModule(console.NewStderrLogger(true))
	conf := rlog.GetDefaultConfig()
	conf.Severity = rlog.SeverityDebug
	rlog.Start(conf)
	defer rlog.Flush()

	//Test all the different log levels
	rlog.Debug("debug log entry")
	rlog.Info("info log entry")
	rlog.Warning("warning log entry")

	// simulate log rotation followed by SIGHUP and then flush.
	err = os.Rename(log_file_name, rotated_log_name)
	if err != nil {
		panic(err)
	}
	rlog.Flush() // will reopen logs

	rlog.Error("error log entry")
	rlog.Fatal("fatal log entry")

	//Generate a couple of IDs and log it
	ids := ""
	for i := 0; i < 10; i++ {
		ids += rlog.GenerateID() + ", "
	}
	rlog.Info("IDs: %s", ids)

	//A deeply nested stack trace
	rlog.Debug("---------------------------")
	level1()
	rlog.Debug("---------------------------")

	//A very long log message
	rlog.Debug(strings.Repeat("hello rlog, ", 1000))

	//All done
	rlog.Debug("Test permutations completed")

}

func level1() {
	level2()
}

func level2() {
	level3()
}

func level3() {
	rlog.Error("nested function call")
}
