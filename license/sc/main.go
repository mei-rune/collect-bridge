package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"license"
	"os"
)

var (
	cmd  = flag.String("cmd", "", "create, import")
	path = flag.String("path", "", "the path of license file.")

	company = flag.String("company", "", "the company name")
	phone   = flag.String("phone", "", "the phone number of contact")
	contact = flag.String("contact", "", "the contact name")
	node    = flag.Int("node", 0, "the node count")
)

func main() {
	flag.Parse()
	args := flag.Args()
	if nil != args {
		switch len(args) {
		case 1:
			flag.Set("cmd", args[0])
		case 2:
			flag.Set("cmd", args[0])
			flag.Set("path", args[1])
		default:
			fmt.Println("arguments is too much.")
			os.Exit(-1)
			return
		}
	}

	switch *cmd {
	case "create":
		create_cmd()
	case "import":
		import_cmd()
	default:
		fmt.Println("unknown command -", *cmd)
		return
	}
}

func create_cmd() {
	if 0 == len(*path) {
		fmt.Println("argument 'path' is empty.")
		os.Exit(-1)
		return
	}

	for _, s := range []string{"company", "phone", "contact"} {
		v := flag.Lookup(s).Value.String()
		if 0 == len(v) {
			fmt.Println("argument '" + s + "' is empty.")
			os.Exit(-1)
			return
		}
	}

	if 0 == *node {
		fmt.Println("argument 'node' is not equal zero.")
		os.Exit(-1)
		return
	}

	lsn := map[string]interface{}{
		"company":    *company,
		"phone":      *phone,
		"contact":    *contact,
		"node":       *node,
		"hd":         license.GetAllHD(),
		"interfaces": license.GetAllInterfaces()}

	var buffer bytes.Buffer
	if e := json.NewEncoder(&buffer).Encode(lsn); nil != e {
		fmt.Println(e.Error())
		os.Exit(-1)
		return
	}

	encryted, e := license.Encrypt(buffer.Bytes())
	if nil != e {
		fmt.Println("encrypt failed,", e.Error())
		os.Exit(-1)
		return
	}

	if e = ioutil.WriteFile(*path, []byte(encryted), 0); nil != e {
		fmt.Println("write tsn to file,", e.Error())
		os.Exit(-1)
		return
	}
}

func import_cmd() {
	if 0 == len(*path) {
		fmt.Println("argument 'path' is empty.")
		os.Exit(-1)
		return
	}
	data, e := ioutil.ReadFile(*path)
	if nil != e {
		fmt.Println("读数据文件失败 -", e)
		os.Exit(-1)
		return
	}

	if nil == data || 0 == len(data) {
		fmt.Println("读数据文件失败 - 内容为空")
		os.Exit(-1)
		return
	}

	_, e = license.DecryptoLicense(data)
	if nil != e {
		fmt.Println("解密对象失败(第一步) -", e)
		os.Exit(-1)
		return
	}

}
