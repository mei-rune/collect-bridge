package sampling

import (
	"commons"
	"commons/netutils"
	"log"
	"sync"
	"time"
)

var is_icmp_test = true

type icmpBucket struct {
	l sync.Mutex
	icmpBuffer
	updated_at int64
	created_at int64
}

func (self *icmpBucket) IsExpired(now int64) bool {
	return now-self.updated_at > *icmp_expired
}

func (self *icmpBucket) IsTimeout(now int64) bool {
	last := self.icmpBuffer.Last()
	if nil == last {
		return now-self.created_at > *icmp_timeout
	}

	return now-last.SampledAt > *icmp_timeout
}

// type pinger interface {
// 	Close()

// 	GetChannel() <-chan *netutils.PingResult
// 	Send(raddr string, echo []byte) error
// }

// type mock_pinger struct {
// 	send int32
// }

// func (self *mock_pinger) Close() {
// }

// func (self *mock_pinger) GetChannel() <-chan *netutils.PingResult {
// 	return nil
// }

// func (self *mock_pinger) Send(raddr string, echo []byte) error {
// 	atomic.AddInt32(&self.send, 1)
// }

type icmpWorker struct {
	l           sync.Mutex
	icmpBuffers map[string]*icmpBucket
	v4          *netutils.Pinger
	c           chan string
}

func (self *icmpWorker) Init() error {
	// if is_icmp_test {
	// 	self.v4 = &mock_pinger{}
	// } else {
	v4, e := netutils.NewPinger("ip4:icmp", "0.0.0.0", nil, 10)
	if nil != e {
		return e
	}
	self.v4 = v4
	//}
	go self.run()
	return nil
}

func (self *icmpWorker) Close() {
	close(self.c)
	if nil != self.v4 {
		self.v4.Close()
	}
	log.Println("[icmp] server is closed")
}

func (self *icmpWorker) OnTick() {
}

func (self *icmpWorker) run() {
	count := uint64(0)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	is_running := true
	for is_running {
		select {
		case <-ticker.C:
			count += 1
			self.scan(int(count % *icmp_poll_interval))
		case s, ok := <-self.c:
			if !ok {
				is_running = false
				break
			}
			self.v4.Send(s, nil)
		case res, ok := <-self.v4.GetChannel():
			if !ok {
				is_running = false
				break
			}

			if nil != res.Err {
				log.Println("[icmp] recv error from pinger,", res.Err)
				break
			}

			self.Push(res)
		}
	}
}

func (self *icmpWorker) clearTimeout() {
	self.l.Lock()
	defer self.l.Unlock()
	expired := make([]string, 0, 10)
	now := time.Now().Unix()
	for k, bucket := range self.icmpBuffers {
		if bucket.IsExpired(now) {
			expired = append(expired, k)
		}
	}

	for _, k := range expired {
		delete(self.icmpBuffers, k)
		log.Println("[icmp] '" + k + "' is expired.")
	}
}

func (self *icmpWorker) scan(c int) {
	if 3 < c {
		return
	}

	self.l.Lock()
	defer self.l.Unlock()
	now := time.Now().Unix()
	expired := make([]string, 0, 10)
	for k, bucket := range self.icmpBuffers {
		if bucket.IsExpired(now) {
			expired = append(expired, k)
			log.Println("[icmp] '" + k + "' with updated_at waw '" + time.Unix(bucket.updated_at, 0).String() + "' is expired.")
		} else {
			if 0 == c || bucket.IsTimeout(now) {
				self.v4.Send(k, nil)
			}
		}
	}

	for _, k := range expired {
		delete(self.icmpBuffers, k)
	}
}

func (self *icmpWorker) Push(res *netutils.PingResult) {
	bucket := self.Get(res.Addr.String())
	if nil == bucket {
		return
	}

	bucket.l.Lock()
	defer bucket.l.Unlock()
	tmp := bucket.BeginPush()
	tmp.SampledAt = res.Timestamp
	tmp.Result = true
	bucket.CommitPush()
}

func (self *icmpWorker) Get(id string) *icmpBucket {
	self.l.Lock()
	defer self.l.Unlock()

	if buffer, ok := self.icmpBuffers[id]; ok {
		return buffer
	}
	return nil
}

func (self *icmpWorker) GetOrCreate(id string) (*icmpBucket, bool) {
	self.l.Lock()
	defer self.l.Unlock()

	if buffer, ok := self.icmpBuffers[id]; ok {
		return buffer, false
	}

	now := time.Now().Unix()
	w := &icmpBucket{created_at: now,
		updated_at: now}
	w.icmpBuffer.init(make([]IcmpResult, *icmp_buffer_size))
	self.icmpBuffers[id] = w

	log.Println("[icmp] add '", id, "' to scan list.")
	return w, true
}

func (self *icmpWorker) Stats() map[string]interface{} {
	self.l.Lock()
	defer self.l.Unlock()

	stats := make([]string, 0, len(self.icmpBuffers))
	for k, _ := range self.icmpBuffers {
		stats = append(stats, k)
	}

	return map[string]interface{}{"name": "icmpWorker", "ports": stats}
}

func (self *icmpWorker) Call(ctx MContext) commons.Result {
	address, e := ctx.GetString("@address")
	if nil != e {
		return commons.ReturnWithIsRequired("address")
	}
	bucket, ok := self.GetOrCreate(address)
	if ok {
		self.c <- address
	}

	bucket.l.Lock()
	defer bucket.l.Unlock()

	now := time.Now().Unix()
	bucket.updated_at = now

	list := bucket.icmpBuffer.All()
	if nil == list || 0 == len(list) {
		if !bucket.IsTimeout(now) {
			return commons.ReturnWithInternalError(pendingError.Error())
		}
		return commons.Return(map[string]interface{}{"result": false, "list": list})
	}

	return commons.Return(map[string]interface{}{"result": !bucket.IsTimeout(now), "list": list})
}

func init() {
	Methods["default_icmp"] = newRouteSpec("get", "icmp", "the result of icmp ping", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			v, ok := params["backgroundWorkers"]
			if !ok {
				return nil, commons.IsRequired("backgroundWorkers")
			}

			bw, ok := v.(BackgroundWorkers)
			if !ok {
				return nil, commons.TypeError("'backgroundWorkers' isn't a BackgroundWorkers")
			}

			drv := &icmpWorker{icmpBuffers: make(map[string]*icmpBucket), c: make(chan string, 10)}
			e := drv.Init()
			if nil != e {
				return nil, e
			}
			bw.Add("icmpWorker", drv)
			return drv, nil
		})
}
