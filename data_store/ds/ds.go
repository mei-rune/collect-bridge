package main

import (
	"data_store"
	"flag"
)

func main() {
	data_store.LicenseUrl = flag.String("license_srv", "http://127.0.0.1:37076/", "")
	data_store.Main()
}
