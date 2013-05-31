package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	root_dir    = flag.String("root", ".", "the root directory")
	config_file = flag.String("config", "./<program_name>.conf", "the config file path")
)

func fileExist(nm string) {
	fs, e := os.Stat(nm)
	if nil != e {
		panic(fmt.Sprintf("[panic] %v", e))
	}
	return !fs.IsDir()
}

func dirExist(nm string) {
	fs, e := os.Stat(nm)
	if nil != e {
		panic(fmt.Sprintf("[panic] %v", e))
	}
	return fs.IsDir()
}

func usage() {
	fmt.Fprint(os.Stderr, `daemon.exe [options] 
Options:
`)
	flag.PrintDefaults()
}

func getDefaultConfigFile() string {

	if "" == *config_file {
		*config_file = "./<program_name>.conf"
	}

	if "./<program_name>.conf" != *config_file {
		return filepath.Clean(filepath.Abs(*config_file))
	}

	program := filepath.Base(os.Args[0])

	nm := filepath.Clean(filepath.Abs(filepath.Join(*root_dir, program+".conf")))
	if fileExist(nm) {
		return nm
	}
	nm = filepath.Clean(filepath.Abs(filepath.Join(*root_dir, "etc", program+".conf")))
	if fileExist(nm) {
		return nm
	}
	nm = filepath.Clean(filepath.Abs(filepath.Join(*root_dir, "conf", program+".conf")))
	if fileExist(nm) {
		return nm
	}

	nm := filepath.Clean(filepath.Abs(filepath.Join(*root_dir, "daemon.conf")))
	if fileExist(nm) {
		return nm
	}
	nm = filepath.Clean(filepath.Abs(filepath.Join(*root_dir, "etc", "daemon.conf")))
	if fileExist(nm) {
		return nm
	}
	nm = filepath.Clean(filepath.Abs(filepath.Join(*root_dir, "conf", "daemon.conf")))
	if fileExist(nm) {
		return nm
	}
	return filepath.Clean(filepath.Abs(filepath.Join(*root_dir, program+".conf")))
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
		return
	}

	if "" == *root_dir {
		*root_dir = filepath.Abs(filepath.Dir(os.Args[0]))
	} else {
		*root_dir = filepath.Abs(*root_dir)
	}
	if dirExist(*root_dir) {
		fmt.Println("root directory '%v' is not exist.", *root_dir)
		return
	}

	e := os.Chdir(*root_dir)
	if nil != e {
		fmt.Println("change current dir to \"" + *root_dir + "\"")
	}

	*config_file = getDefaultConfigFile()
	if fileExist(*config_file) {
		fmt.Println("config file '%v' is not exist.", *config_file)
		return
	}

	procs, e := readConfigs(*root_dir, *config_file)
	if nil != e {
		fmt.Println("read config file failed, %v", e)
		return
	}

  
}
