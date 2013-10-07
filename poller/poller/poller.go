package main

import (
	"flag"
	"poller"
)

func main() {
	poller.LicenseUrl = flag.String("license_srv", "http://127.0.0.1:37076/", "")
	poller.Main()
}
