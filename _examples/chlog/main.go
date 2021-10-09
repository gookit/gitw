package main

import (
	"fmt"

	"github.com/gookit/gitwrap"
	"github.com/gookit/gitwrap/chlog"
	"github.com/gookit/goutil"
)

func main() {
	cl := chlog.New()
	cl.WithConfig(func(c *chlog.Changelog) {
		// some settings ...
	})

	logCmd := gitwrap.New("log", "--reverse")
	logCmd.Add("--no-merges")
	logCmd.Addf("--pretty=format:\"%s\"", cl.LogFormat)
	logCmd.Addf("%s...%s", "v0.1.0", "HEAD")

	fmt.Println("CMD>", logCmd.Cmdline());

	logOut := logCmd.SafeOutput()

	cl.SetLogText(logOut)

	goutil.PanicIfErr(cl.Generate())

	fmt.Println(cl.Changelog())
}
