/*
Application loggerObjects are test permutations testing rlog logObjects.
*/
package main

import (
	"github.com/rightscale/rlog"
	"github.com/rightscale/rlog/stdout"
)

func main() {
	rlog.EnableModule(stdout.NewStdoutLogger(true))
	rlog.Start(rlog.GetDefaultConfig())
	defer rlog.Flush()

	rlog.Info("================= Rlog log objects test permutations =================")

	//Create log object
	logger := rlog.NewLogger()

	//Test all the different log levels
	logger.Debug("debug log entry")
	logger.Info("info log entry")
	logger.Error("error log entry")
	logger.Fatal("fatal log entry")

	//Generate an ID
	logger.Info("ID 1: %s", logger.GenerateID())
	logger.Info("ID 2: %s", logger.GenerateID())

	//A deeply nested stack trace
	rlog.Info("------------- Testing stack trace --------------")
	level1()

	//All done
	rlog.Info("================= Test permutations completed =================")
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
