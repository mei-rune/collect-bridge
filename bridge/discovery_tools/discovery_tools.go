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
)

var (
	depth       = flag.Int("depth", 5, "the depth")
	timeout     = flag.Int("timeout", 5, "the timeout")
	network     = flag.String("ip-range", "", "the ip range")
	communities = flag.String("communities", "public;public1", "the community")
	proxy       = flag.String("proxy", "127.0.0.1:7070", "the address of bridge proxy")
	dbproxy     = flag.String("dbproxy", "127.0.0.1:7071", "the address of mdb proxy")
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
	return commons.GetReturn(result).(string), nil
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
	body := readAll(resp.Body)
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(body), &result)

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

func save(drv interface{}) (string, error) {
	js, err := json.Marshal(drv)
	if nil != err {
		return "", fmt.Errorf("marshal device to json failed, %s", err.Error())
	}

	resp, e := http.Post("http://"+*dbproxy+"/mdb/device", "application/json", bytes.NewBuffer([]byte(js)))
	if nil != e {
		return "", fmt.Errorf("create device failed, %v", e)
	}
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("create device failed, %v, %v", resp.StatusCode, readAll(resp.Body))
	}
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return "", fmt.Errorf("create device failed, %v", e)
	}
	return commons.GetReturn(result).(string), nil
}
func main() {

	flag.Parse()
	targets := flag.Args()
	if nil == targets || 0 != len(targets) {
		flag.Usage()
		return
	}

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
		fmt.Println("======================")
		e := json.NewEncoder(os.Stdout).Encode(res)
		if nil != e {
			fmt.Println(e)
		}
		fmt.Println("======================")

		devices, ok := res.(map[string]interface{})
		if ok {

			for k, _ := range devices {
				fmt.Println(k)
			}

			for k, drv := range devices {
				_, e := save(drv)
				if nil != e {
					fmt.Println(k, e)
				}
			}
		} else {
			fmt.Println("this is not a array")
		}
	}

	err = discoveryDelete(id)
	if nil != err {
		fmt.Println(err)
		return
	}
}
