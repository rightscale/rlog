rlog introduction
=================

rlog is an advanced logging library for GO supporting various output modules, log levels, filtering
and tagging. The API is very similar to the API of the GO log standard library but differs because
the API of the standard logging library cannot carry enough information to support advanced features
found to be useful in larger production environments.

How to contribute
-----------------
Bug fixes, code improvements and additional output modules are always welcome.


Usage
=====

rlog is comprised of a core and various output modules (each implemented in its own package).

Example: stdout and syslog
--------------------------

First, use EnableModule() to activate the syslog logger and the stdout logger (log output is written
to both sources). Next, initialize the rlog core by calling Start() and passing a configuration. The
deferred flush call ensures that output is written before the application terminates.

	func main(){
		syslogModule, err := syslog.NewLocalSyslogLogger()
		if err != nil {
			panic("Getting syslog logger instance failed")
		}

		rlog.EnableModule(syslogModule)
		rlog.EnableModule(stdout.NewStdoutLogger(true))
		rlog.Start(rlog.GetDefaultConfig())
		defer rlog.Flush()

		//Write your code here
	}

Architecture & details
======================

This section describes how rlog works internally and interacts with the various output modules. This
is useful for people who want to understand the internals of rlog or would like to contribute to it.

Core
----

The core implements the user facing API, message formatting, stack trace retrieval (depending on the
severity of the log message) and submission to the various registered output modules. The basic
processing steps:

	1. User submits log message through rlog API
	2. Drop message if tag or log level criteria match
	3. Retrieve and cut stack trace if log message is error or fatal
	4. Forward message to each module

Output modules
--------------

To be compatible with the rlog core, all output modules must implement the "rlogModule" interface
providing a message channel and a flush command channel. Log messages are sent through the message
channel using the message format defined in common/RlogMsg.

Each output module is launched by the rlog core upon a call rlog.Start(...) and runs in its own
goroutine. This isolates the output module from the rlog core. So even if an IO operation takes a
long time, the caller of rlog will not be effected by this.

Furthermore, all channel operations performed by the rlog core are non-blocking. This ensures that
even if an output module crashes (e.g. not accepting log messages anymore), the rlog caller (your
application) is not blocked rlog.

Writing an output module
------------------------

To write an rlog module, the following steps are required:

	1. Create a new package and give it an expressive, short name
	2. Write a "LaunchModule" function to satisfy the "rlogModule" interface
	3. Import package "common" to obtain access to the log message type
	4. Write a constructor returning an instance of the log module. Use arguments to obtain configuration.

Running the unit tests
----------------------

Tests for rlog are written using the "gocheck" testing framework (providing fixtures, etc.). Before
running "go test", please fetch the following additional library:

	go get launchpad.net/gocheck

License
=======
The MIT License (MIT)

Copyright (c) 2013 brsc

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

Maintained by
 - [Sapphire Team](https://wookiee.rightscale.com/display/rightscale/Meet+the+Sapphire+Team)

Merge to master whitelist
 - @ryanwilliamson
