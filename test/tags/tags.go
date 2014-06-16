/*
Application tags are test permutations testing tag based filtering in rlog.
*/
package main

import (
	"github.com/rightscale/rlog"
	"github.com/rightscale/rlog/console"
)

func main() {

	const TAG1 string = "tag1"
	const TAG2 string = "tags"

	rlog.EnableModule(console.NewStdoutLogger(true))
	conf := rlog.GetDefaultConfig()
	conf.DisableTagsExcept([]string{TAG1})
	rlog.Start(conf)

	rlog.InfoT(TAG1, "This msg should appear")
	rlog.InfoT(TAG2, "This msg should NOT appear")

	defer rlog.Flush()
}
