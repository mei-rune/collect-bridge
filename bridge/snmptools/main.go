package main

import (
	"code.google.com/p/mahonia"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"unicode"
	//"unicode/utf16"
	"errors"
	"unicode/utf8"
)

type VB struct {
	Oid   string
	Value string
}

var (
	proxy           = flag.String("proxy", "127.0.0.1:7070", "the address of proxy server, default: 127.0.0.1:7070")
	target          = flag.String("target", "127.0.0.1,161", "the address of snmp agent, default: 127.0.0.1,161")
	community       = flag.String("community", "public", "the community of snmp agent, default: public")
	action          = flag.String("action", "walk", "the action, default: walk")
	version         = flag.String("version", "2c", "the version of snmp protocal, default: 2c")
	secret_name     = flag.String("name", "", "the name, default: \"\"")
	auth_passphrase = flag.String("auth", "", "the auth passphrase, default: \"\"")
	priv_passphrase = flag.String("priv", "", "the priv passphrase, default: \"\"")
	started_oid     = flag.String("oid", "1.3.6", "the start oid, default: 1.3.6")
	from_charset    = flag.String("charset", "GB18030", "the charset of octet string, default: GB18030")
	help            = flag.Bool("h", false, "print help")

	decoder mahonia.Decoder
	out     io.Writer
)

func IsAsciiAndPrintable(bytes []byte) bool {
	for _, c := range bytes {
		if c >= unicode.MaxASCII {
			return false
		}

		if !unicode.IsPrint(rune(c)) {
			return false
		}
	}
	return true
}

func IsUtf8AndPrintable(bytes []byte) bool {
	for 0 != len(bytes) {
		c, l := utf8.DecodeRune(bytes)
		if utf8.RuneError == c {
			return false
		}

		if !unicode.IsPrint(c) {
			return false
		}
		bytes = bytes[l:]
	}
	return true
}

func IsUtf16AndPrintable(bytes []byte) bool {
	if 0 != len(bytes)%2 {
		return false
	}

	for i := 0; i < len(bytes); i += 2 {
		u16 := binary.LittleEndian.Uint16(bytes[i:])
		if !unicode.IsPrint(rune(u16)) {
			return false
		}
	}
	return true
}

func IsUtf32AndPrintable(bytes []byte) bool {
	if 0 != len(bytes)%4 {
		return false
	}

	for i := 0; i < len(bytes); i += 4 {
		u32 := binary.LittleEndian.Uint32(bytes[i:])
		if !unicode.IsPrint(rune(u32)) {
			return false
		}
	}
	return true
}

func printValue(value string) {

	if !strings.HasPrefix(value, "[octets") {
		fmt.Println(value)
		return
	}

	bytes, err := hex.DecodeString(value[8:])
	if nil != err {
		fmt.Println(value)
		return
	}

	if nil != decoder {
		fmt.Println(value)

		for 0 != len(bytes) {
			c, length, status := decoder(bytes)
			switch status {
			case mahonia.SUCCESS:
				if unicode.IsPrint(c) {
					out.Write(bytes[0:length])
				} else {
					for i := 0; i < length; i++ {
						out.Write([]byte{'.'})
					}
				}
				bytes = bytes[length:]
			case mahonia.INVALID_CHAR:
				out.Write([]byte{'.'})
				bytes = bytes[1:]
			case mahonia.NO_ROOM:
				out.Write([]byte{'.'})
				bytes = bytes[0:0]
			case mahonia.STATE_ONLY:
				bytes = bytes[length:]
			}
		}
		out.Write([]byte{'\n'})
		return
	}

	if IsUtf8AndPrintable(bytes) {
		fmt.Println(string(bytes))
		return
	}

	if IsUtf16AndPrintable(bytes) {
		rr := make([]rune, len(bytes)/2)
		for i := 0; i < len(bytes); i += 2 {
			rr[i/2] = rune(binary.LittleEndian.Uint16(bytes[i:]))
		}
		fmt.Println(string(rr))
		return
	}
	if IsUtf32AndPrintable(bytes) {
		rr := make([]rune, len(bytes)/4)
		for i := 0; i < len(bytes); i += 4 {
			rr[i/4] = rune(binary.LittleEndian.Uint32(bytes[i:]))
		}
		fmt.Println(string(rr))
		return
	}
	fmt.Println(value)

	for _, c := range bytes {
		if c >= unicode.MaxASCII {
			fmt.Print(".")
		} else {
			fmt.Print(string(c))
		}
	}
	fmt.Println()
}

func main() {

	out = os.Stdout
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	if "guess" != *from_charset {
		decoder = mahonia.NewDecoder(*from_charset)
	}

	switch *action {
	case "walk":
		walk()
	case "next":
		next()
	case "get":
		get()
	case "table":
		table()
	case "sys", "system":
		oid := "1.3.6.1.2.1.1"
		started_oid = &oid
		table()
	case "sys.descr", "sys.description", "system.descr", "system.description":
		oid := "1.3.6.1.2.1.1.1.0"
		started_oid = &oid
		get()
	case "interface", "interfaces":
		oid := "1.3.6.1.2.1.2.2.1"
		started_oid = &oid
		table()
	case "arp":
		oid := "1.3.6.1.2.1.4.22.1"
		started_oid = &oid
		table()
	case "ip":
		oid := "1.3.6.1.2.1.4.20.1"
		started_oid = &oid
		table()
	case "mac":
		oid := "1.3.6.1.2.1.4.20.1" // ?
		started_oid = &oid
		table()
	case "route":
		oid := "1.3.6.1.2.1.4.21.1"
		started_oid = &oid
		table()
	default:
		fmt.Println("unsupported action - " + *action)
	}
}

func get() {
	_, err := invoke("get", *started_oid)
	if nil != err {
		fmt.Println(err.Error())
	}
}

func next() {
	_, err := invoke("next", *started_oid)
	if nil != err {
		fmt.Println(err.Error())
	}
}

func walk() {
	var err error = nil
	oid := *started_oid
	for {
		oid, err = invoke("next", oid)
		if nil != err {
			fmt.Println(err.Error())
			break
		}
	}
}
func table() {
	var err error = nil
	oid := *started_oid
	for {
		oid, err = invoke("next", oid)
		if nil != err {
			fmt.Println(err.Error())
			break
		}

		if !strings.HasPrefix(oid, *started_oid) {
			break
		}
	}
}

func createUrl(action, oid string) (string, error) {

	var url string
	switch *version {
	case "2", "2c", "v2", "v2c", "1", "v1":
		url = fmt.Sprintf("http://%s/snmp/"+action+"/%s/%s?community=%s", *proxy, *target, strings.Replace(oid, ".", "_", -1), *community)
	case "3", "v3":
		url = fmt.Sprintf("http://%s/snmp/"+action+"/%s/%s?version=3&secmodel=usm&secname=%s", *proxy, *target, strings.Replace(oid, ".", "_", -1), *secret_name)
		if "" != *auth_passphrase {
			url = url + "&auth_pass=" + *auth_passphrase
			if "" != *priv_passphrase {
				url = url + "&priv_pass=" + *priv_passphrase
			}
		}
	default:
		return "", errors.New("version is error.")
	}
	return url, nil
}

func invoke(action, oid string) (string, error) {
	var err error

	url, err := createUrl(action, oid)
	if nil != err {
		return "", err
	}

	fmt.Println("Get " + url)
	resp, err := http.Get(url)
	if nil != err {
		return "", fmt.Errorf("get failed - " + err.Error())
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return "", fmt.Errorf("read body failed - " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(bytes))
	}

	var vbs map[string]string
	err = json.Unmarshal(bytes, &vbs)
	if nil != err {
		return "", errors.New("unmarshal failed - " + err.Error() + "\n" + string(bytes))
	}
	if 0 == len(vbs) {
		return "", errors.New("result is empty." + "\n" + string(bytes))
	}

	err = nil
	var next_oid string
	for key, value := range vbs {

		if strings.HasPrefix(value, "[error") {
			if !strings.HasPrefix(value, "[error:11]") {
				err = fmt.Errorf("invalid value - %v", value)
			} else {
				err = errors.New("walk end.")
			}
			return "", err
		}

		next_oid = key
		printValue(value)
	}

	return next_oid, nil
}
