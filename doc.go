/*
Package rlog is an advanced and extensible logger for Go.

Package rlog implements the core logging facility, which are the user API and log message
processing. The rlog output modules are producing the output, i.e. without any enabled modules, rlog
does not produce any output. Note that all methods provided by rlog are thread safe.

Configuring rlog & enabling modules

Output modules offer a "new" method to create a new instance for that particular output type and rlog
offers the EnableModule method to enable each method satisfying the required interface provided by
the rlog. rlog is configured by retrieving and modifying the default configuration using the
GetDefaultConfig() method. Once started, rlog's configuration cannot be modified. rlog is
usually initialized in main. When calling "rlog.Start()", it is advisable to call "defer
rlog.Flush() right after to ensure that upon termination of the main method, all log entries are
written.

Example setup procedure with stdout and syslog output:

	syslogModule, err := syslog.NewLocalSyslogLogger()
	if err != nil {
		panic("Getting syslog logger instance failed")
	}

	rlog.EnableModule(syslogModule)
	rlog.EnableModule(stdout.NewStdoutLogger(true))
	rlog.Start(rlog.GetDefaultConfig())
	defer rlog.Flush()

Example: setup using tags

	const TAG1 string = "tag1"
	const TAG2 string = "tags"

	//Start rlog, display only msg carrying TAG1
	rlog.EnableModule(stdout.NewStdoutLogger(true))
	conf := rlog.GetDefaultConfig()
	conf.DisableTagsExcept([]string{TAG1})
	rlog.Start(conf)
	defer rlog.Flush()

	//Output tagged messages
	rlog.InfoT(TAG1, "This msg appears")
	rlog.InfoT(TAG2, "This msg does NOT appear")

Producing log output

rlog exists as a singleton and output can be produced by simply importing the rlog package and
calling the various print messages on it. Output methods ending with a T (e.g. infoT) require a tag
argument. Tags are user defined strings. It is highly recommended to define tags as constants to
avoid typos. Tags are currently NOT implemented.

Example:

	rlog.Debug("debug log entry")
	rlog.Info("info log entry")
	rlog.Warn("warn log entry")
	rlog.ErrorT(DATABASE, "Connection terminated")
	rlog.Fatal("fatal log entry")

Generating IDs

GenerateID() generates a unique, hex formatted string ID. The initial value is random and each successive call
increments it.

Log objects

Log objects are can be retrieved using the NewLogger method. The user gets back an object referring to the singleton
logger, offering the same API as the rlog package. This allows to mock rlog package using an interface requirement
when generating shared libraries.
*/
package rlog
