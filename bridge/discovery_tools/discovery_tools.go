package main

import (
	"bytes"
	"commons"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"web"
)

var (
	depth       = flag.Int("depth", 5, "the depth")
	timeout     = flag.Int("timeout", 5, "the timeout")
	network     = flag.String("ip-range", "", "the ip range")
	communities = flag.String("communities", "public", "the community")
	proxy       = flag.String("proxy", "127.0.0.1:7070", "the address of proxy")

	address   = flag.String("http", ":7071", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
)

func HttpPut(endpoint string, bodyType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("PUT", endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return http.DefaultClient.Do(req)
}
func httpDelete(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func mainHandle(rw *web.Context) {
	errFile := "_log_/error.html"
	_, err := os.Stat(errFile)
	if err == nil || os.IsExist(err) {
		content, _ := ioutil.ReadFile(errFile)
		rw.WriteString(string(content))
		return
	}
	rw.WriteString("Hello, World!")
}

func readAll(r io.Reader) string {
	bs, e := ioutil.ReadAll(r)
	if nil != e {
		panic(e.Error())
	}
	return string(bs)
}

func discoveryCreate(params map[string]interface{}) (string, error) {
	js, err := json.Marshal(params)
	if nil != err {
		return "", fmt.Errorf("marshal params to json failed, %s", err.Error())
	}

	resp, e := http.Post("http://"+*proxy+"/discovery", "application/json", bytes.NewBuffer([]byte(js)))
	if nil != e {
		return "", fmt.Errorf("create discovery failed, %v", e)
	}
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("create discovery failed, %v, %v", resp.StatusCode, readAll(resp.Body))
	}
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return "", fmt.Errorf("create discovery failed, %v", e)
	}
	return result["id"].(string), nil
}

func discoveryGet(id string, params map[string]string) (interface{}, error) {
	url := "http://" + *proxy + "/discovery/" + id + "?"
	for k, v := range params {
		url += (k + "=" + v + "&")
	}
	resp, e := http.Get(url[:len(url)-1])
	if nil != e {
		return "", fmt.Errorf("get discovery message failed, %v", e)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("get discovery message failed, %v, %v", resp.StatusCode, readAll(resp.Body))
	}
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return "", fmt.Errorf("get discovery message failed, %v", e)
	}

	return commons.GetReturn(result), nil
}
func discoveryDelete(id string) error {
	resp, e := httpDelete("http://" + *proxy + "/discovery/" + id)
	if nil != e {
		return fmt.Errorf("delete discovery failed, %v", e)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("delete discovery failed, %v, %v", resp.StatusCode, readAll(resp.Body))
	}
	return nil
}

func main() {

	flag.Parse()
	targets := flag.Args()
	if nil == targets || 0 != len(targets) {
		flag.Usage()
		return
	}

	svr := web.NewServer()
	svr.Config.Name = "meijing-discovery v1.0"
	svr.Config.Address = *address
	svr.Config.StaticDirectory = *directory
	svr.Config.CookieSecret = *cookies
	svr.Get("/", mainHandle)

	params := map[string]interface{}{}

	communities2 := strings.Split(*communities, ";")
	if nil != communities2 && 0 != len(communities2) {
		params["communities"] = communities2
	}

	network2 := strings.Split(*network, ";")
	if nil != network2 && 0 != len(network2) {
		params["ip-range"] = network2
	}
	params["depth"] = *depth

	id, err := discoveryCreate(params)
	if nil != err {
		fmt.Println(err)
		return
	}

	go svr.Run()
	for {
		values, err := discoveryGet(id, map[string]string{"dst": "message"})
		if nil != err {
			fmt.Println(err)
			return
		}
		messages, ok := values.([]interface{})
		if ok {
			isEnd := false
			for _, msg := range messages {
				if msg == "end" {
					isEnd = true
				}
				fmt.Println(msg)
			}
			if isEnd {
				break
			}
		} else {
			fmt.Println(values)
		}
	}
	res, err := discoveryGet(id, map[string]string{})
	if nil != err {
		fmt.Println(err)
		return
	}
	if nil != res {
		bytes, e := json.MarshalIndent(res, "", "  ")
		if nil != e {
			fmt.Println(e)
			return
		}
		fmt.Println(string(bytes))
	}

	err = discoveryDelete(id)
	if nil != err {
		fmt.Println(err)
		return
	}
}
