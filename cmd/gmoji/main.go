package main

import (
	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/errorx"
)

var cmd = cflag.New(func(c *cflag.CFlags) {
	c.Version = "1.0.0"
	c.Desc = "quick show or render git emoji code"
})

var opts = struct {
	render string
	search string
}{}

func main() {
	cmd.StringVar(&opts.render, "render", "", "want rendered text;;r")
	cmd.StringVar(&opts.search, "search", "", "want rendered text;;s")
	cmd.Func = execute
}

func execute(c *cflag.CFlags) error {
	return errorx.Raw("TODO")
}
