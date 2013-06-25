package ds

import (
	"bytes"
	"errors"
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
	refresh     time.Duration
}

type caches_request struct {
	action int
	ch     chan *caches_request
	target string
	cache  *Cache
	e      error
}

func NewCaches(refresh time.Duration, client *Client, alias map[string]string) *Caches {
	alias_names := alias
	if nil == alias_names {
		alias_names = map[string]string{}
	}

	caches := &Caches{
		caches:      make(map[string]*Cache),
		ch:          make(chan *caches_request, 5),
		client:      client,
		refresh:     refresh,
		alias_names: alias_names}
	go caches.serve()
	return caches
}

func (c *Caches) GetCache(target string) (*Cache, error) {
	if 0 == len(target) {
		return nil, errors.New("target is empty.")
	}

	ch := make(chan *caches_request)
	defer close(ch)

	if target_name, ok := c.alias_names[target]; ok {
		c.ch <- &caches_request{action: GET, ch: ch, target: target_name}
	} else {
		c.ch <- &caches_request{action: GET, ch: ch, target: target}
	}

	resp := <-ch
	return resp.cache, resp.e
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
			break
		}

		if INTERRUPT == req.action {
			if nil != req.ch {
				req.ch <- req
			}
			break
		}
		c.doCommand(req)
	}
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
		if e.Code() == TABLE_NOT_EXISTS {
			return nil, nil
		}
		return nil, e
	}
	return NewCache(c.refresh, c.client, target), nil
}
