package main

import (
	"flag"
	"web"
)

var (
	address   = flag.String("http", ":7070", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
)

func main() {
	flag.Parse()
	svr := web.NewServer()
	svr.Config.Name = "meijing-bridge v1.0"
	svr.Config.Address = *address
	svr.Config.StaticDirectory = *directory
	svr.Config.CookieSecret = *cookies

	registerSNMP(svr)

	svr.Run()
}
