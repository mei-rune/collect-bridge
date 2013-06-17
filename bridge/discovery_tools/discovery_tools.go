package main

import (
	"bytes"
	"commons"
	"ds"
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
	timeout     = flag.Int("timeout", 10, "the timeout")
	network     = flag.String("ip-range", "", "the ip range")
	communities = flag.String("communities", "public;public1", "the community")
	proxy       = flag.String("proxy", "127.0.0.1:7070", "the address of bridge proxy")
	dbproxy     = flag.String("dbproxy", "127.0.0.1:7071", "the address of ds proxy")
)

func httpPut(endpoint string, bodyType string, body io.Reader) (*http.Response, error) {
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
	result := commons.SimpleResult{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return "", fmt.Errorf("create discovery failed, %v", e)
	}
	return result.Value().AsString()
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
	result := commons.SimpleResult{}
	e = json.Unmarshal([]byte(body), &result)
	if nil != e {
		return "", fmt.Errorf("get discovery message failed, %v", e)
	}

	return result.InterfaceValue(), nil
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

func findDeviceByAddress(address string) (map[string]interface{}, error) {
	res, e := findDevice(map[string]string{"@address": address})
	if nil == res {
		return nil, e
	}
	if 0 == len(res) {
		return nil, fmt.Errorf("device is not found - %s", address)
	}
	return res[0], nil
}
func findDevice(params map[string]string) ([]map[string]interface{}, error) {
	url := "http://" + *dbproxy + "/mdb/device/query" + "?"
	for k, v := range params {
		url += (k + "=" + v + "&")
	}

	resp, e := http.Get(url[:len(url)-1])
	if nil != e {
		return nil, fmt.Errorf("get device message failed, %v", e)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get device message failed, %v, %v", resp.StatusCode, readAll(resp.Body))
	}
	body := readAll(resp.Body)
	result := commons.SimpleResult{}
	e = json.Unmarshal([]byte(body), &result)

	if nil != e {
		return nil, fmt.Errorf("get device message failed, %v", e)
	}

	return result.Value().AsObjects()
}

func save(attributes map[string]interface{}) (string, string, error) {
	client := ds.NewClient("http://" + *dbproxy)

	id, e := client.SaveBy("device", map[string]string{"address": fmt.Sprint(attributes["address"])}, attributes)
	if nil != e {
		return "", "", e
	}

	if nil != client.Warnings {
		fmt.Println(client.Warnings)
	}

	return id, "unknow", nil
}

func createTriggers(attributes map[string]interface{}) []map[string]interface{} {
	return []map[string]interface{}{map[string]interface{}{"type": "metric_trigger",
		"expression": "@every 15s",
		"name":       "cpu-" + fmt.Sprint(attributes["address"]),
		"metric":     "cpu",
		"$action":    []interface{}{map[string]interface{}{"type": "redis_command", "name": "redis-cpu-" + fmt.Sprint(attributes["address"]), "command": "SET", "arg0": "cpu:" + fmt.Sprint(attributes["address"])}}},
		map[string]interface{}{"type": "metric_trigger",
			"expression": "@every 15s",
			"name":       "mem-" + fmt.Sprint(attributes["address"]),
			"metric":     "mem",
			"$action":    []interface{}{map[string]interface{}{"type": "redis_command", "name": "redis-mem-" + fmt.Sprint(attributes["address"]), "command": "SET", "arg0": "mem:" + fmt.Sprint(attributes["address"])}}},
		map[string]interface{}{"type": "metric_trigger",
			"expression": "@every 15s",
			"name":       "interface-" + fmt.Sprint(attributes["address"]),
			"metric":     "interface",
			"$action":    []interface{}{map[string]interface{}{"type": "redis_command", "name": "redis-interface-" + fmt.Sprint(attributes["address"]), "command": "SET", "arg0": "interface:" + fmt.Sprint(attributes["address"])}}}}
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
	params["timeout"] = *timeout
	params["is_read_local"] = true

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
			for k, drv := range devices {
				attributes := drv.(map[string]interface{})
				attributes["$trigger"] = createTriggers(attributes)
				_, action, e := save(attributes)
				if nil != e {
					fmt.Println(action, k, e)
				} else {
					fmt.Println(action, k)
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
