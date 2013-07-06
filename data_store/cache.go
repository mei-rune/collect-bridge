package data_store

import (
	"bytes"
	"commons"
	"fmt"
	"log"
	"runtime"
	"time"
)

const (
	GET        = 0
	CHILDREN   = 1
	DELETE     = 2
	ONLYFIND   = 3
	SET        = 4
	REFRESH    = 5
	REFRESH_OK = 6
	INTERRUPT  = 7 // interrupt
)

type cache_request struct {
	action         int
	id             string
	child_type     string
	child_matchers map[string]commons.Matcher
	ch             chan *cache_request

	snapshots map[string]*RecordVersion

	ok      bool
	result  map[string]interface{}
	results []map[string]interface{}
	e       error
}

type Cache struct {
	objects map[string]map[string]interface{}
	ch      chan *cache_request

	is_refreshing bool
	ticker        *time.Ticker
	client        *Client
	target        string
	includes      string
}

func NewCache(refresh time.Duration, client *Client, target string) *Cache {
	return NewCacheWithIncludes(refresh, client, target, "")
}

func NewCacheWithIncludes(refresh time.Duration, client *Client, target, includes string) *Cache {
	if refresh < 10*time.Second {
		refresh = 10 * time.Second
	}

	cache := &Cache{
		objects:  make(map[string]map[string]interface{}),
		ch:       make(chan *cache_request, 5),
		ticker:   time.NewTicker(refresh),
		client:   client,
		target:   target,
		includes: includes}
	go cache.serve()
	return cache
}

func (c *Cache) Close() {
	ch := make(chan *cache_request)
	defer close(ch)

	c.ch <- &cache_request{
		action: INTERRUPT,
		ch:     ch}

	<-ch

	return
}

var emptyParams = map[string]string{}

func (c *Cache) LoadAll() ([]map[string]interface{}, error) {
	return c.client.FindByWithIncludes(c.target, emptyParams, c.includes)
}

func (c *Cache) Refresh() error {
	snapshots, e := c.client.Snapshot(c.target, emptyParams)
	if nil != e {
		return e
	}

	c.ch <- &cache_request{
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

	_, updated, deleted := Diff(snapshots, old_snapshots)
	for _, n := range [][]string{updated, deleted} {
		if nil == n {
			continue
		}

		for _, id := range n {
			delete(c.objects, id)
		}
	}
}

func (c *Cache) Set(id string, res map[string]interface{}) {
	c.ch <- &cache_request{
		action: SET,
		id:     id,
		result: res}
}

func (c *Cache) Delete(id string) {
	c.ch <- &cache_request{
		action: DELETE,
		id:     id}
}

func (c *Cache) Get(id string) (map[string]interface{}, error) {
	ch := make(chan *cache_request)
	defer close(ch)

	c.ch <- &cache_request{
		action: GET,
		ch:     ch,
		id:     id}
	select {
	case r := <-ch:
		return r.result, r.e
	case <-time.After(30 * time.Second):
		return nil, commons.TimeoutErr
	}
}

func (c *Cache) GetChildren(id, child_type string, matchers map[string]commons.Matcher) ([]map[string]interface{}, error) {
	ch := make(chan *cache_request)
	defer close(ch)

	c.ch <- &cache_request{
		action:         CHILDREN,
		child_type:     child_type,
		child_matchers: matchers,
		ch:             ch,
		id:             id}
	select {
	case r := <-ch:
		return r.results, r.e
	case <-time.After(30 * time.Second):
		return nil, commons.TimeoutErr
	}
}

func (c *Cache) Find(id string) map[string]interface{} {
	ch := make(chan *cache_request)
	defer close(ch)

	c.ch <- &cache_request{
		action: ONLYFIND,
		ch:     ch,
		id:     id}

	select {
	case r := <-ch:
		return r.result
	case <-time.After(30 * time.Second):
		return nil
	}
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
					defer func() { c.ch <- &cache_request{action: REFRESH_OK} }()
					c.Refresh()
				}()
			}
		case req := <-c.ch:
			if nil == req {
				goto end
			}
			if INTERRUPT == req.action {
				if nil != req.ch {
					req.ch <- req
				}
				goto end
			}
			c.doCommand(req)
		}
	}
end:
	fmt.Println("exited while recv a close command.")
}
func (c *Cache) doChildren(req *cache_request) *cache_request {
	if res, ok := c.objects[req.id]; ok {
		req.result = res
	} else {
		req.result, req.e = c.client.FindByIdWithIncludes(c.target, req.id, c.includes)
		if nil == req.e {
			c.set(req.id, req.result)
		}
	}
	// res := req.result["$"+req.child_type]
	// if nil != res {
	// 	if result, ok := res.(map[string]interface{}); ok {
	// 		if nil == req.child_matchers || commons.IsMatch(result, req.child_matchers) {
	// 			req.results = []map[string]interface{}{result}
	// 		}
	// 	} else if results, ok := res.([]interface{}); ok {
	// 		for _, v := range results {
	// 			if result, ok := v.(map[string]interface{}); ok {
	// 				if nil == req.child_matchers || commons.IsMatch(result, req.child_matchers) {
	// 					req.results = append(req.results, result)
	// 				}
	// 			}
	// 		}
	// 	} else if results, ok := res.([]map[string]interface{}); ok {
	// 		for _, result := range results {
	// 			if nil == req.child_matchers || commons.IsMatch(result, req.child_matchers) {
	// 				req.results = append(req.results, result)
	// 			}
	// 		}
	// 	}
	// }

	req.results = GetChildrenForm(req.result["$"+req.child_type], req.child_matchers)
	return req
}

func (c *Cache) doCommand(req *cache_request) {
	switch req.action {
	case REFRESH_OK:
		c.is_refreshing = false
	case REFRESH:
		c.compare(req.snapshots)
	case GET:
		if res, ok := c.objects[req.id]; ok {
			req.result = res
		} else {
			req.result, req.e = c.client.FindByIdWithIncludes(c.target, req.id, c.includes)
			if nil == req.e {
				c.set(req.id, req.result)
			}
		}
		req.ch <- req

	case CHILDREN:
		req.ch <- c.doChildren(req)
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
	default:
		if nil != req.ch {
			req.e = fmt.Errorf("unsupported command - %v", req.action)
			req.ch <- req
		}
	}
}
