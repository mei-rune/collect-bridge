package sampling

import (
	"commons"
	"github.com/runner-mei/snmpclient"
	"log"
	"strings"
	"sync"
	"time"
)

type snmpBucket struct {
	byAddress  *bucketByAddress
	snmpBuffer snmpTestResultBuffer
	updated_at int64
	created_at int64

	address   string
	version   snmpclient.SnmpVersion
	community string
}

func (self *snmpBucket) IsExpired(now int64) bool {
	return now-self.updated_at > *snmp_test_expired
}

func (self *snmpBucket) IsTimeout(now int64) bool {
	last := self.snmpBuffer.Last()
	if nil == last {
		return now-self.created_at > *snmp_test_timeout
	}

	return now-last.SampledAt > *snmp_test_timeout
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

type testingRequst struct {
	id  int
	key string
}

type bucketByAddress struct {
	l           sync.Mutex
	snmpBuffers map[string]*snmpBucket
	pendings    []*testingRequst
	next_id     int
}

type snmpWorker struct {
	l       sync.Mutex
	buckets map[string]*bucketByAddress
	v4      *snmpclient.Pinger
	c       chan *snmpBucket
}

func (self *snmpWorker) Init() error {
	v4, e := snmpclient.NewPinger("udp4", ":0", 10)
	if nil != e {
		return e
	}
	self.v4 = v4
	go self.run()
	return nil
}

func (self *snmpWorker) Close() {
	close(self.c)
	self.v4.Close()
}

func (self *snmpWorker) OnTick() {
}

func (self *snmpWorker) run() {
	count := uint64(0)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	is_running := true
	for is_running {
		select {
		case <-ticker.C:
			self.scan(int(count % *snmp_test_poll_interval))
		case bucket, ok := <-self.c:
			if !ok {
				is_running = false
				break
			}

			self.send(bucket.version.String()+"/"+bucket.community, bucket)
		case res, ok := <-self.v4.GetChannel():
			if !ok {
				is_running = false
				break
			}

			if nil != res.Error {
				break
			}

			self.Push(res)
		}
	}
}

func (self *snmpWorker) send(key string, bucket *snmpBucket) {
	// Send(id int, raddr string, version SnmpVersion, community string) error {
	bucket.byAddress.next_id++
	if e := self.v4.Send(bucket.byAddress.next_id, bucket.address, bucket.version, bucket.community); nil != e {
		bucket.byAddress.pendings = append(bucket.byAddress.pendings, &testingRequst{id: bucket.byAddress.next_id, key: key})
	}
}

func (self *snmpWorker) clearTimeout() {
	self.l.Lock()
	defer self.l.Unlock()
	now := time.Now().Unix()

	deleted := make([]string, 0, 10)
	for address, byAddress := range self.buckets {

		expired := make([]string, 0, 10)
		for k, bucket := range byAddress.snmpBuffers {
			if bucket.IsExpired(now) {
				expired = append(expired, k)
			}
		}

		for _, k := range expired {
			delete(byAddress.snmpBuffers, k)
			log.Println("[snmp] '" + address + "/" + k + "' is expired.")
		}

		if 0 == len(byAddress.snmpBuffers) {
			deleted = append(deleted, address)
		}
	}

	for _, address := range deleted {
		delete(self.buckets, address)
		log.Println("[snmp] '" + address + "' is expired.")
	}
}

func (self *snmpWorker) scan(c int) {
	if 3 < c {
		return
	}

	self.l.Lock()
	defer self.l.Unlock()
	now := time.Now().Unix()

	deleted := make([]string, 0, 10)
	for address, byAddress := range self.buckets {
		expired := make([]string, 0, 10)
		for k, bucket := range byAddress.snmpBuffers {
			if bucket.IsExpired(now) {
				expired = append(expired, k)
			} else {
				if 0 == c || bucket.IsTimeout(now) {
					self.send(k, bucket)
				}
			}
		}

		for _, k := range expired {
			delete(byAddress.snmpBuffers, k)
			log.Println("[snmp] '" + address + "/" + k + "' is expired.")
		}

		if 0 == len(byAddress.snmpBuffers) {
			deleted = append(deleted, address)
		}
	}

	for _, address := range deleted {
		delete(self.buckets, address)
		log.Println("[snmp] '" + address + "' is expired.")
	}
}

func (self *snmpWorker) Push(res *snmpclient.PingResult) {
	address := res.Addr.String()
	idx := strings.IndexRune(address, ':')
	if -1 != idx {
		address = address[:idx]
	}

	byAddress := self.Get(address)
	if nil == byAddress {
		return
	}

	byAddress.l.Lock()
	defer byAddress.l.Unlock()

	if nil == byAddress.pendings {
		return
	}

	var testing *testingRequst
	for _, pending := range byAddress.pendings {
		if pending.id == res.Id {
			testing = pending
		}
	}

	if nil == testing {
		return
	}

	bucket, ok := byAddress.snmpBuffers[testing.key]
	if !ok {
		return
	}

	tmp := bucket.snmpBuffer.BeginPush()
	tmp.SampledAt = res.Timestamp.Unix()
	tmp.Result = true
	bucket.snmpBuffer.CommitPush()
}

func (self *snmpWorker) Get(address string) *bucketByAddress {
	self.l.Lock()
	defer self.l.Unlock()

	byAddress, ok := self.buckets[address]
	if !ok {
		return nil
	}
	return byAddress
}

func (self *snmpWorker) GetOrCreate(address string, version snmpclient.SnmpVersion, community string) (*snmpBucket, bool) {
	self.l.Lock()
	defer self.l.Unlock()

	byAddress, ok := self.buckets[address]
	if !ok {
		byAddress = &bucketByAddress{snmpBuffers: make(map[string]*snmpBucket)}
		self.buckets[address] = byAddress
	}

	id := version.String() + "/" + community
	if buffer, ok := byAddress.snmpBuffers[id]; ok {
		return buffer, false
	}

	w := &snmpBucket{byAddress: byAddress,
		created_at: time.Now().Unix(),
		address:    address,
		version:    version,
		community:  community}

	w.snmpBuffer.init(make([]SnmpTestResult, *snmp_test_buffer_size))
	byAddress.snmpBuffers[id] = w
	return w, true
}

func (self *snmpWorker) Stats() map[string]interface{} {
	self.l.Lock()
	defer self.l.Unlock()

	stats := make([]map[string]interface{}, 0, len(self.buckets))
	for _, byAddress := range self.buckets {
		for _, bucket := range byAddress.snmpBuffers {
			stats = append(stats, map[string]interface{}{"address": bucket.address, "version": bucket.version, "community": bucket.community})
		}
	}

	return map[string]interface{}{"name": "snmpWorker", "ports": stats}
}

func (self *snmpWorker) Call(ctx MContext) commons.Result {
	address, e := ctx.GetString("@address")
	if nil != e {
		return commons.ReturnWithIsRequired("address")
	}
	port := ctx.GetStringWithDefault("snmp.port", "")
	if 0 != len(port) {
		if strings.HasPrefix(port, ":") {
			address += port
		} else {
			address = address + ":" + port
		}
	}
	version := ctx.GetStringWithDefault("snmp.version", "")
	if 0 == len(version) {
		return commons.ReturnWithBadRequest(snmpNotExistsError.Error())
	}
	community := ctx.GetStringWithDefault("snmp.read_community", "")
	if 0 == len(community) {
		return commons.ReturnWithIsRequired("snmp.read_community")
	}

	version_int, e := snmpclient.ParseSnmpVersion(version)
	if nil != e {
		return commons.ReturnWithBadRequest(e.Error())
	}

	if version_int == snmpclient.SNMP_V3 {
		// snmp_params["snmp.secmodel"] = params.GetStringWithDefault("snmp.sec_model", "")
		// snmp_params["snmp.auth_pass"] = params.GetStringWithDefault("snmp."+rw+"_auth_pass", "")
		// snmp_params["snmp.priv_pass"] = params.GetStringWithDefault("snmp."+rw+"_priv_pass", "")
		// snmp_params["snmp.max_msg_size"] = params.GetStringWithDefault("snmp.max_msg_size", "")
		// snmp_params["snmp.context_name"] = params.GetStringWithDefault("snmp.context_name", "")
		// snmp_params["snmp.identifier"] = params.GetStringWithDefault("snmp.identifier", "")
		// snmp_params["snmp.engine_id"] = params.GetStringWithDefault("snmp.engine_id", "")
		return commons.ReturnWithBadRequest(snmpv3UnsupportedError.Error())
	}

	bucket, ok := self.GetOrCreate(address, version_int, community)
	if ok {
		self.c <- bucket
	}

	bucket.byAddress.l.Lock()
	defer bucket.byAddress.l.Unlock()

	now := time.Now().Unix()
	bucket.updated_at = now

	return commons.Return(map[string]interface{}{"result": bucket.IsTimeout(now), "list": bucket.snmpBuffer.All()})
}

func init() {
	Methods["default_snmp_test"] = newRouteSpec("get", "snmp_test", "the result of snmp ping", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			v, ok := params["backgroundWorkers"]
			if !ok {
				return nil, commons.IsRequired("backgroundWorkers")
			}

			bw, ok := v.(BackgroundWorkers)
			if !ok {
				return nil, commons.TypeError("'backgroundWorkers' isn't a BackgroundWorkers")
			}

			drv := &snmpWorker{buckets: make(map[string]*bucketByAddress), c: make(chan *snmpBucket, 10)}
			e := drv.Init()
			if nil != e {
				return nil, e
			}
			bw.Add("snmpWorker", drv)
			return drv, nil
		})
}
