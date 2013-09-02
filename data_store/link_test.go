package data_store

import (
	"commons/types"
	"testing"
)

var old_link = map[string]interface{}{"name": "dd",
	"custom_speed_up":   12,
	"custom_speed_down": 12,
	"description":       "",
	"from_device":       0,
	"from_if_index":     1,
	"to_device":         0,
	"to_if_index":       1,
	"link_type":         12,
	"forward":           true,
	"from_based":        true}

func copyFrom(from, addition map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range from {
		res[k] = v
	}

	if nil != addition {
		for k, v := range addition {
			res[k] = v
		}
	}
	return res
}

func TestLink(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_link", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		if "" == id1 {
			return
		}

		link := copyFrom(old_link, map[string]interface{}{"from_device": id1, "to_device": id2})
		link_id := CreateItForTest(t, client, "network_link", link)

		new_link := findById(t, client, "network_link", link_id)
		b, ok := new_link["forward"]
		if !ok || nil == b {
			t.Error("forward is not exists.")
		} else if b == false {
			t.Error("forward is false.")
		}
		b, ok = new_link["from_based"]
		if !ok || nil == b {
			t.Error("from_based is not exists.")
		} else if b == false {
			t.Error("from_based is false.")
		}
	})
}

func TestLink3(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_link", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		if "" == id1 {
			return
		}

		link := copyFrom(old_link, map[string]interface{}{"from_device": id1, "to_device": id2, "forward": false, "from_based": false})
		link_id := CreateItForTest(t, client, "network_link", link)

		new_link := findById(t, client, "managed_object", link_id)
		b, ok := new_link["forward"]
		if !ok || nil == b {
			t.Error("forward is not exists.")
		} else if b == true {
			t.Error("forward is true.")
		}
		b, ok = new_link["from_based"]
		if !ok || nil == b {
			t.Error("forward is not exists.")
		} else if b == true {
			t.Error("forward is true.")
		}
	})
}
