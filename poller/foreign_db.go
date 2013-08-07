package poller

import (
	"bytes"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type data_object struct {
	c          chan error
	attributes map[string]interface{}
}

type foreignDb struct {
	name       string
	action     string
	url        string
	c          chan *data_object
	status     int32
	wait       sync.WaitGroup
	last_error *expvar.String
}

func (self *foreignDb) isRunning() bool {
	return 1 == atomic.LoadInt32(&self.status)
}

func (self *foreignDb) Close() {
	atomic.StoreInt32(&self.status, 0)
	self.wait.Wait()
}

func (self *foreignDb) run() {
	is_running := int32(1)
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
			msg := buffer.String()
			self.last_error.Set(msg)
			log.Println(msg)

			if !self.isRunning() {
				os.Exit(-1)
			}
		}

		log.Println("foreignDb is exit.")
		close(self.c)
		atomic.StoreInt32(&is_running, 0)
		self.wait.Done()
	}()

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		defer func() {
			if o := recover(); nil != o {
				log.Println("[panic]", o)
			}
			ticker.Stop()
			self.wait.Done()
		}()

		<-ticker.C
		for 1 == atomic.LoadInt32(&is_running) {
			self.c <- nil
			<-ticker.C
		}
	}()

	self.wait.Add(1)

	objects := make([]*data_object, 0, 10)
	for self.isRunning() {
		self.runOne(objects[0:0], 10)
	}
}

func (self *foreignDb) runOne(cached_array []*data_object, max_size int) {
	objects := self.recvObjects(cached_array, max_size)
	if 0 == len(objects) {
		// idle
		return
	}

	var e error
	failed := make([]error, len(objects))
	buffer := bytes.NewBuffer(make([]byte, 0, 1024*(1+len(objects))))
	buffer.WriteByte('[')

	encoder := json.NewEncoder(buffer)

	for i, obj := range objects {
		remain := buffer.Len()
		if 0 != i {
			buffer.WriteByte(',')
		}

		e = encoder.Encode(obj.attributes)
		if nil != e {
			failed[i] = e
			buffer.Truncate(remain)
		}
	}
	buffer.WriteByte(']')

	e = self.save(buffer)

	for i, obj := range objects {
		if nil == obj.c {
			continue
		}

		if nil != failed[i] {
			self.reply(obj.c, failed[i])
		} else {
			self.reply(obj.c, e)
		}
	}
}

func (self *foreignDb) reply(c chan<- error, e error) {
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			msg := buffer.String()
			log.Println(msg)
		}
	}()
	c <- e
}

func (self *foreignDb) recvObjects(objects []*data_object, max_size int) []*data_object {
	cmd := <-self.c
	if nil == cmd {
		return objects
	}

	objects = append(objects, cmd)
	for self.isRunning() {
		select {
		case cmd := <-self.c:
			if nil == cmd {
				continue
			}

			objects = append(objects, cmd)
			if max_size < len(objects) {
				return objects
			}
		default:
			return objects
		}
	}
	return objects
}

func (self *foreignDb) save(objects *bytes.Buffer) (e error) {
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
			msg := buffer.String()
			self.last_error.Set(msg)
			e = errors.New(msg)
		}

		if nil != e {
			self.last_error.Set(e.Error())
		}
	}()

	//fmt.Println(self.action, self.url, objects.String())
	req, err := http.NewRequest(self.action, self.url, objects)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return e
	}

	// Install closing the request body (if any)
	defer func() {
		if nil != resp.Body {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != 200 {
		resp_body, e := ioutil.ReadAll(resp.Body)
		if nil != e {
			return e
		}

		if nil == resp_body || 0 == len(resp_body) {
			return fmt.Errorf("%v: error", resp.StatusCode)
		}

		return fmt.Errorf("%v: %v", resp.StatusCode, string(resp_body))
	}

	return nil
}

func newForeignDb(name, url string) (*foreignDb, error) {
	var varString *expvar.String = nil

	varE := expvar.Get("foreign_db." + name)
	if nil != varE {
		varString, _ = varE.(*expvar.String)
		if nil == varString {
			varString = expvar.NewString("foreign_db." + name + "." + time.Now().String())
		}
	} else {
		varString = expvar.NewString("foreign_db." + name)
	}

	db := &foreignDb{name: name,
		action:     "PUT",
		url:        url,
		c:          make(chan *data_object, 3000),
		status:     1,
		last_error: varString}
	go db.run()
	db.wait.Add(1)
	return db, nil
}
