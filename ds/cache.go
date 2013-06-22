package ds

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"time"
)

const (
	GET        = 0
	DELETE     = 1
	ONLYFIND   = 2
	SET        = 3
	REFRESH    = 4
	REFRESH_OK = 5
)

type request struct {
	action int
	id     string
	ch     chan *request

	snapshots map[string]*RecordVersion

	ok     bool
	result map[string]interface{}
	e      error
}

type Cache struct {
	objects map[string]map[string]interface{}
	ch      chan *request

	is_refreshing bool
	ticker        *time.Ticker
	client        *Client
	target        string
}

func NewCache(refresh time.Duration, client *Client, target string) *Cache {
	cache := &Cache{
		objects: make(map[string]map[string]interface{}),
		ch:      make(chan *request, 5),
		ticker:  time.NewTicker(refresh),
		client:  client,
		target:  target}
	go cache.serve()
	return cache
}

var emptyParams = map[string]string{}

func (c *Cache) LoadAll() ([]map[string]interface{}, error) {
	return c.client.FindBy(c.target, emptyParams)
}

func (c *Cache) Refresh() error {
	snapshots, e := c.client.Snapshot(c.target, emptyParams)
	if nil != e {
		return e
	}

	c.ch <- &request{
		action:    REFRESH,
		snapshots: snapshots}
	return nil
}

func (c *Cache) compare(snapshots map[string]*RecordVersion) {
	old_snapshots := make(map[string]*RecordVersion, len(c.objects))
	for id, result := range c.objects {
		version, _ := GetRecordVersionFrom(result)
		old_snapshots[id] = version
	}

	for id, version := range snapshots {
		old_version, ok := old_snapshots[id]
		if !ok {
			//fmt.Println("not found, skip", id)
			continue
		}
		delete(old_snapshots, id)
		if nil == version {
			//fmt.Println("version is nil, skip", id)
			continue
		}
		if nil == old_version {
			//fmt.Println("old version is nil, reload", id)
			delete(c.objects, id)
			continue
		}

		if version.UpdatedAt.After(old_version.UpdatedAt) {
			//fmt.Println("after, reload", id)
			delete(c.objects, id)
		} // else {
		//	fmt.Println("not after, skip", id, version.UpdatedAt, old_version.UpdatedAt)
		//}
	}

	for id, _ := range old_snapshots {
		delete(c.objects, id)
		// fmt.Println("delete", id)
	}
}

func (c *Cache) Set(id string, res map[string]interface{}) {
	c.ch <- &request{
		action: SET,
		id:     id,
		result: res}
}

func (c *Cache) Delete(id string) {
	c.ch <- &request{
		action: DELETE,
		id:     id}
}

func (c *Cache) Get(id string) (map[string]interface{}, error) {
	ch := make(chan *request)
	defer close(ch)

	c.ch <- &request{
		action: GET,
		ch:     ch,
		id:     id}
	r := <-ch

	return r.result, r.e
}

func (c *Cache) Find(id string) map[string]interface{} {
	ch := make(chan *request)
	defer close(ch)

	c.ch <- &request{
		action: ONLYFIND,
		ch:     ch,
		id:     id}
	r := <-ch

	return r.result
}

func (c *Cache) set(id string, res map[string]interface{}) {
	//fmt.Println("cached", id)
	c.objects[id] = res
}

func (c *Cache) serve() {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			log.Print(buffer.String())
		}
		close(c.ch)
		c.ticker.Stop()
	}()

	for {
		select {
		case <-c.ticker.C:
			if !c.is_refreshing {
				c.is_refreshing = true
				go func() {
					defer func() { c.ch <- &request{action: REFRESH_OK} }()
					c.Refresh()
				}()
			}
		case req := <-c.ch:
			if nil == req {
				break
			}
			c.doCommand(req)
		}
	}
}

func (c *Cache) doCommand(req *request) {
	switch req.action {
	case REFRESH_OK:
		c.is_refreshing = false
	case REFRESH:
		c.compare(req.snapshots)
	case GET:
		if res, ok := c.objects[req.id]; ok {
			req.result = res
		} else {
			req.result, req.e = c.client.FindById(c.target, req.id)
			if nil == req.e {
				c.set(req.id, req.result)
			}
		}
		req.ch <- req
	case SET:
		c.objects[req.id] = req.result
		if nil != req.ch {
			req.ch <- req
		}
	case DELETE:
		if nil == req.ch {
			delete(c.objects, req.id)
		} else {
			if _, ok := c.objects[req.id]; ok {
				delete(c.objects, req.id)
				req.ok = true
			} else {
				req.ok = false
			}
			req.ch <- req
		}

	case ONLYFIND:
		if res, ok := c.objects[req.id]; ok {
			req.result = res
		} else {
			req.result = nil
		}
		req.ch <- req
	}
}
