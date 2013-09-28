package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"license"
	"os"
	"path/filepath"
)

var (
	cmd  = flag.String("cmd", "", "create, import")
	path = flag.String("path", "", "the path of license file.")

	company = flag.String("company", "", "the company name")
	phone   = flag.String("phone", "", "the phone number of contact")
	contact = flag.String("contact", "", "the contact name")
	node    = flag.Int("node", 0, "the node count")
)

func abs(pa string) string {
	s, e := filepath.Abs(pa)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func searchDir() (string, bool) {
	files := []string{abs(filepath.Join("lib")),
		abs(filepath.Join("..", "lib"))}
	for _, file := range files {
		if st, e := os.Stat(file); nil == e && nil != st && st.IsDir() {
			return filepath.Base(file), true
		}
	}
	return abs("."), false
}

func main() {
	flag.Parse()
	args := flag.Args()
	if nil != args && 0 != len(args) {
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

	lic_txt, e := license.DecryptoLicense(data)
	if nil != e {
		fmt.Println(e)
		os.Exit(-1)
		return
	}
	var attributes map[string]interface{}
	decoder := json.NewDecoder(bytes.NewBuffer(lic_txt))
	decoder.UseNumber()
	if e = decoder.Decode(&attributes); nil != e {
		fmt.Println("解析数据失败(第二步) -" + e.Error())
		os.Exit(-1)
		return
	}

	for _, s := range []string{"company", "phone", "contact", "node"} {
		v, ok := attributes[s]
		if !ok || nil == v {
			fmt.Println("'" + s + "' 不存在 -" + e.Error())
			os.Exit(-1)
			return
		}
		sv := fmt.Sprint(v)
		if "" == sv {
			fmt.Println("'" + s + "' 不存在 -" + e.Error())
			os.Exit(-1)
			return
		}
		fmt.Println(s+":", sv)
	}

	pa, _ := searchDir()
	nm := filepath.Join(pa, "tpt.lic")
	if e = ioutil.WriteFile(nm, data, 0); nil != e {
		fmt.Println("写文件 '"+nm+"'失败 -", e)
		os.Exit(-1)
		return
	}
}
