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
	"unicode/utf8"
)

type VB struct {
	Oid   string
	Value string
}

var (
	proxy        = flag.String("proxy", "127.0.0.1:7070", "the address of proxy server, default: 127.0.0.1:7070")
	target       = flag.String("target", "127.0.0.1,161", "the address of snmp agent, default: 127.0.0.1,161")
	community    = flag.String("community", "public", "the community of snmp agent, default: public")
	started_oid  = flag.String("oid", "1.3.6", "the start oid, default: 1.3.6")
	from_charset = flag.String("charset", "GB18030", "the charset of octet string, default: GB18030")
	help         = flag.Bool("h", false, "print help")

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

	oid := *started_oid
	for {
		url := fmt.Sprintf("http://%s/snmp/next/%s/%s?community=%s", *proxy, *target, strings.Replace(oid, ".", "_", -1), *community)
		fmt.Println("Get " + url)
		resp, err := http.Get(url)
		if nil != err {
			fmt.Println("get failed - " + err.Error())
			break
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			fmt.Println("read body failed - " + err.Error())
			break
		}
		var vbs map[string]string
		err = json.Unmarshal(bytes, &vbs)
		if nil != err {
			fmt.Println(string(bytes))
			fmt.Println("unmarshal failed - " + err.Error())
			break
		}
		if 0 == len(vbs) {
			fmt.Println(string(bytes))
			fmt.Println("result is empty.")
			break
		}

		isFailed := false
		for key, value := range vbs {

			if strings.HasPrefix(value, "[error") {
				if !strings.HasPrefix(value, "[error:11]") {
					fmt.Println(value)
					fmt.Println("invalid value.")
				} else {
					fmt.Println("walk end.")
				}
				isFailed = true
				break
			}

			oid = key
			printValue(value)
		}

		if isFailed {
			break
		}
	}
}
