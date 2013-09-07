package sampling

import (
	"commons"
	"sync"
)

type icmpBucket struct {
	l sync.Mutex
	icmpBuffer
}

type icmpWorker struct {
	l           sync.Mutex
	icmpBuffers map[string]*icmpBucket
}

func (self *icmpWorker) Init() error {
	return nil
}

func (self *icmpWorker) Close() {
}

func (self *icmpWorker) OnTick() {
}

func (self *icmpWorker) Get(id string) *icmpBucket {
	self.l.Lock()
	defer self.l.Unlock()

	if buffer, ok := self.icmpBuffers[id]; ok {
		return buffer
	}

	w := &icmpBucket{}
	w.icmpBuffer.init(make([]IcmpResult, *icmp_buffer_size))
	self.icmpBuffers[id] = w
	return w
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
	//ttl := ctx.GetIntWithDefault("@ttl", 255)
	bucket := self.Get(address)
	bucket.l.Lock()
	defer bucket.l.Unlock()
	return commons.Return(bucket.icmpBuffer.All())
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

			drv := &icmpWorker{icmpBuffers: make(map[string]*icmpBucket)}
			e := drv.Init()
			if nil != e {
				return nil, e
			}
			bw.Add("icmpWorker", drv)
			return drv, nil
		})
}
