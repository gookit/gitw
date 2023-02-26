package main

import (
	"fmt"

	"github.com/gookit/gitw/gmoji"
	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/errorx"
)

var cmd = cflag.New(func(c *cflag.CFlags) {
	c.Version = "1.0.0"
	c.Desc = "quick show or render git emoji code"
})

var geOpts = struct {
	list   bool
	lang   string
	render string
	search cflag.String
}{}

// quick run:
//
//	go run ./cmd/gmoji
//	go run ./cmd/gmoji -h
//
// install to GOPATH/bin:
//
//	go install ./cmd/gmoji
func main() {
	cmd.Var(&geOpts.search, "search", "search emoji by keywords,multi by comma;;s")
	cmd.StringVar(&geOpts.render, "render", "", "want rendered text;;r")
	cmd.StringVar(&geOpts.lang, "lang", gmoji.LangEN, "select language for emojis;;L")
	cmd.BoolVar(&geOpts.list, "list", false, "list all git emojis;;ls,l")

	cmd.Func = execute
	cmd.QuickRun()
}

func execute(c *cflag.CFlags) error {
	em, err := gmoji.Emojis(geOpts.lang)
	if err != nil {
		return err
	}

	if geOpts.list {
		cliutil.Warnf("All git emojis(total: %d):\n", em.Len())
		fmt.Println(em.String())
		return nil
	}

	if geOpts.search != "" {
		sub := em.Search(geOpts.search.Strings(), 10)

		cliutil.Warnf("Matched emojis(total: %d):\n", sub.Len())
		fmt.Println(sub.String())
		return nil
	}

	return errorx.Raw("TODO render ...")
}
