/*
Application modules are test permutations testing various rlog output modules.
*/
package main

import (
  "github.com/brsc/rlog"
  "github.com/brsc/rlog/console"
  "github.com/brsc/rlog/file"
  "github.com/brsc/rlog/syslog"
  "strings"
)

func main() {

  //Setup syslog module
  syslogModule, err := syslog.NewLocalSyslogLogger()
  if err != nil {
    panic("Getting syslog logger instance failed")
  }

  //Setup file logger
  fileModule, err := file.NewFileLogger("test.txt", true, true)
  if err != nil {
    panic("Getting file logger instance failed: " + err.Error())
  }

  rlog.EnableModule(syslogModule)
  rlog.EnableModule(fileModule)
  rlog.EnableModule(console.NewStdoutLogger(true))
  rlog.Start(rlog.GetDefaultConfig())
  defer rlog.Flush()

  //Test all the different log levels
  rlog.Debug("debug log entry")
  rlog.Info("info log entry")
  rlog.Warn("warn log entry")
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
  rlog.Debug(strings.Repeat("hello Rightscale", 1000))

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
