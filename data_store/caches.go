package data_store

import (
	"bytes"
	"commons"
	"fmt"
	"log"
	"runtime"
	"time"
)

type Caches struct {
	alias_names map[string]string
	caches      map[string]*Cache
	ch          chan *caches_request
	client      *Client
	includes    string
	refresh     time.Duration
}

type caches_request struct {
	action int
	ch     chan *caches_request
	target string
	cache  *Cache
	e      error
}

func NewCaches(refresh time.Duration, client *Client, includes string, alias map[string]string) *Caches {
	alias_names := alias
	if nil == alias_names {
		alias_names = map[string]string{}
	}

	caches := &Caches{
		caches:      make(map[string]*Cache),
		ch:          make(chan *caches_request, 5),
		client:      client,
		includes:    includes,
		refresh:     refresh,
		alias_names: alias_names}
	go caches.serve()
	return caches
}

func (c *Caches) GetCache(target string) (*Cache, error) {
	if 0 == len(target) {
		return nil, commons.ParameterIsEmpty
	}

	ch := make(chan *caches_request)
	defer close(ch)

	if target_name, ok := c.alias_names[target]; ok {
		c.ch <- &caches_request{action: GET, ch: ch, target: target_name}
	} else {
		c.ch <- &caches_request{action: GET, ch: ch, target: target}
	}
	select {
	case resp := <-ch:
		return resp.cache, resp.e
	case <-time.After(30 * time.Second):
		return nil, commons.TimeoutErr
	}
}

func (c *Caches) Close() {
	ch := make(chan *caches_request)
	defer close(ch)

	c.ch <- &caches_request{action: INTERRUPT, ch: ch}
	<-ch
	return
}

func (c *Caches) serve() {
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
	}()

	for {
		req := <-c.ch
		if nil == req {
			c.clearCache()
			break
		}

		if INTERRUPT == req.action {
			c.clearCache()
			if nil != req.ch {
				req.ch <- req
			}
			break
		}
		c.doCommand(req)
	}
}

func (c *Caches) clearCache() {
	for _, cache := range c.caches {
		if nil != cache {
			cache.Close()
		}
	}
	c.caches = make(map[string]*Cache)
}

func (c *Caches) doCommand(req *caches_request) {
	switch req.action {
	case GET:
		cache, ok := c.caches[req.target]
		if !ok {
			var e error
			cache, e = c.createCache(req.target)
			if nil != e {
				req.e = e
			} else {
				c.caches[req.target] = cache
			}
		}
		req.cache = cache
		req.ch <- req
	default:
		if nil != req.ch {
			req.e = fmt.Errorf("unsupported command - %v", req.action)
			req.ch <- req
		}
	}
}

func (c *Caches) createCache(target string) (*Cache, error) {
	_, e := c.client.Count(target, emptyParams)
	if nil != e {
		if commons.TableIsNotExists == e.Code() {
			return nil, nil
		}
		return nil, e
	}
	return NewCacheWithIncludes(c.refresh, c.client, target, c.includes), nil
}
