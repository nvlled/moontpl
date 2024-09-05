package main

import (
	"github.com/nvlled/moontpl"
)

func main() {
	moontpl.SetModule("mymod", moontpl.ModMap{
		"foo": func() { println("foo called") },
		"bar": func() { println("bar called") },
	})
	moontpl.ExecuteCLI()
}
