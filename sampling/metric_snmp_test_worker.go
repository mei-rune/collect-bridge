package sampling

import (
	"commons"
	"fmt"
	"github.com/runner-mei/snmpclient"
	"log"
	"net"
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
	//elapsed   time.Duration
}

func (self *snmpBucket) IsExpired(now int64) bool {
	return now-self.updated_at > *snmp_test_expired
}

func (self *snmpBucket) IsTimeout(now int64) bool {
	last := self.snmpBuffer.Last()
	if nil == last {
		return now-self.created_at > *snmp_test_timeout
	}

	return now-last.RecvAt > *snmp_test_timeout
}

type testingRequst struct {
	id        int
	key       string
	timestamp time.Time
}

type bucketByAddress struct {
	ra          *net.UDPAddr
	l           sync.Mutex
	snmpBuffers map[string]*snmpBucket
	pendings    snmpPendingBuffer
	next_id     int

	//elapsed time.Duration
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
	if nil != self.v4 {
		self.v4.Close()
	}
	log.Println("[snmp_test] server is closed")
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
			count += 1
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
				log.Println("[snmp_test] recv error from pinger,", res.Error)
				break
			}

			//log.Println(res.Addr.String())

			self.Push(res)
		}
	}
}

func (self *snmpWorker) send(key string, bucket *snmpBucket) {
	bucket.byAddress.next_id++
	if e := self.v4.SendPdu(bucket.byAddress.next_id, bucket.byAddress.ra, bucket.version, bucket.community); nil != e {
		log.Println("[snmp_test] send pdu to", bucket.address, "with version is", bucket.version,
			"and community is", bucket.community, "failed, ", e)
	} else {
		bucket.byAddress.pendings.Push(testingRequst{id: bucket.byAddress.next_id, key: key, timestamp: time.Now()})
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
				log.Println("[snmp_test] '" + address + "/" + k + "' with updated_at waw '" + time.Unix(bucket.updated_at, 0).String() + "' is expired.")
			}
		}

		for _, k := range expired {
			delete(byAddress.snmpBuffers, k)
		}

		if 0 == len(byAddress.snmpBuffers) {
			deleted = append(deleted, address)
		}
	}

	for _, address := range deleted {
		delete(self.buckets, address)
		log.Println("[snmp_test] '" + address + "' is expired.")
	}
}

func (self *snmpWorker) scan(c int) {
	if 3 < c {
		return
	}

	self.l.Lock()
	defer self.l.Unlock()
	start_at := time.Now()
	now := start_at.Unix()

	deleted := make([]string, 0, 10)
	for address, byAddress := range self.buckets {
		expired := make([]string, 0, 10)
		for k, bucket := range byAddress.snmpBuffers {
			if bucket.IsExpired(now) {
				expired = append(expired, k)
			} else {
				if 0 == c || bucket.IsTimeout(now) {
					//start_s_at := time.Now()
					self.send(k, bucket)
					//bucket.elapsed = time.Now().Sub(start_s_at)
				}
			}
		}

		for _, k := range expired {
			delete(byAddress.snmpBuffers, k)
			log.Println("[snmp_test] '" + address + "/" + k + "' is expired.")
		}

		if 0 == len(byAddress.snmpBuffers) {
			deleted = append(deleted, address)
		}
		//byAddress.elapsed = time.Now().Sub(start_a_at)
	}

	for _, address := range deleted {
		delete(self.buckets, address)
		log.Println("[snmp_test] '" + address + "' is expired.")
	}

	// if 0 == c {
	// 	log.Println("[snmp_test] scan all address - ", time.Now().Sub(start_at).String())
	// 	if 1 < time.Now().Sub(start_at).Seconds() {
	// 		for address, byAddress := range self.buckets {
	// 			log.Println("[snmp_test] elapsed for ", address, " = ", byAddress.elapsed)
	// 			for k, bucket := range byAddress.snmpBuffers {
	// 				log.Println("[snmp_test] elapsed for "+k+" = ", bucket.elapsed)
	// 			}
	// 		}
	// 	}
	// }
}

func (self *snmpWorker) Push(res *snmpclient.PingResult) {
	address := res.Addr.String()
	byAddress := self.Get(address)
	if nil == byAddress {
		log.Println("[snmp_test]", address, "is not exists.")
		return
	}

	byAddress.l.Lock()
	defer byAddress.l.Unlock()

	if byAddress.pendings.IsEmpty() {
		log.Println("[snmp_test]", address, "is not pending, pendings is empty.")
		return
	}

	var testing testingRequst
	for i := 0; i < byAddress.pendings.Size(); i++ {
		pending := byAddress.pendings.Get(i)
		if pending.id == res.Id {
			testing = pending
			byAddress.pendings.RemoveFirst(i + 1)
			break
		}
	}

	if 0 == testing.id {
		log.Println("[snmp_test]", address, "is not pending, it is not in list - id of pid is", res.Id, ", current id is", byAddress.next_id, ".")
		return
	}

	bucket, ok := byAddress.snmpBuffers[testing.key]
	if !ok {
		log.Println("[snmp_test]", address, " and ", testing.key, " is not pending, it may expired.")
		return
	}

	bucket.snmpBuffer.Push(SnmpTestResult{Result: true,
		SendAt:  testing.timestamp.Unix(),
		RecvAt:  res.Timestamp.Unix(),
		Elapsed: res.Timestamp.Sub(testing.timestamp).Nanoseconds() / int64(time.Microsecond) / int64(time.Millisecond)})

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

func (self *snmpWorker) GetOrCreate(address string, version snmpclient.SnmpVersion,
	community string) (*snmpBucket, bool) {
	self.l.Lock()
	defer self.l.Unlock()

	byAddress, ok := self.buckets[address]
	if !ok {
		ra, err := net.ResolveUDPAddr("udp4", address)
		if err != nil {
			panic(fmt.Sprintf("ResolveIPAddr(udp4, %q) failed: %v", address, err))
		}

		byAddress = &bucketByAddress{snmpBuffers: make(map[string]*snmpBucket), ra: ra}
		byAddress.pendings.Init(make([]testingRequst, 10))
		self.buckets[address] = byAddress
	}

	id := version.String() + "/" + community
	if buffer, ok := byAddress.snmpBuffers[id]; ok {
		return buffer, false
	}

	now := time.Now().Unix()
	w := &snmpBucket{byAddress: byAddress,
		created_at: now,
		updated_at: now,
		address:    address,
		version:    version,
		community:  community}

	w.snmpBuffer.init(make([]SnmpTestResult, *snmp_test_buffer_size))
	byAddress.snmpBuffers[id] = w
	log.Println("[snmp_test] add '" + address + "' and " + id + "' to scan list.")
	return w, true
}

func (self *snmpWorker) Stats() map[string]interface{} {
	self.l.Lock()
	defer self.l.Unlock()

	stats := make([]map[string]interface{}, 0, len(self.buckets))
	for _, byAddress := range self.buckets {
		for _, bucket := range byAddress.snmpBuffers {
			stats = append(stats, map[string]interface{}{"address": bucket.address,
				"version": bucket.version, "community": bucket.community})
		}
	}

	return map[string]interface{}{"name": "snmpWorker", "ports": stats}
}

func (self *snmpWorker) Call(ctx MContext) (interface{}, error) {
	address, e := ctx.GetString("@address")
	if nil != e {
		return nil, IsRequired("address")
	}
	port := ctx.GetStringWithDefault("snmp.port", "")
	if 0 != len(port) {
		if strings.HasPrefix(port, ":") {
			address += port
		} else {
			address = address + ":" + port
		}
	} else {
		address = address + ":161"
	}
	version := ctx.GetStringWithDefault("snmp.version", "")
	if 0 == len(version) {
		return nil, BadRequest(snmpNotExistsError.Error())
	}
	community := ctx.GetStringWithDefault("snmp.read_community", "")
	if 0 == len(community) {
		return nil, IsRequired("snmp.read_community")
	}

	version_int, e := snmpclient.ParseSnmpVersion(version)
	if nil != e {
		return nil, BadRequest(e.Error())
	}

	if version_int == snmpclient.SNMP_V3 {
		// snmp_params["snmp.secmodel"] = params.GetStringWithDefault("snmp.sec_model", "")
		// snmp_params["snmp.auth_pass"] = params.GetStringWithDefault("snmp."+rw+"_auth_pass", "")
		// snmp_params["snmp.priv_pass"] = params.GetStringWithDefault("snmp."+rw+"_priv_pass", "")
		// snmp_params["snmp.max_msg_size"] = params.GetStringWithDefault("snmp.max_msg_size", "")
		// snmp_params["snmp.context_name"] = params.GetStringWithDefault("snmp.context_name", "")
		// snmp_params["snmp.identifier"] = params.GetStringWithDefault("snmp.identifier", "")
		// snmp_params["snmp.engine_id"] = params.GetStringWithDefault("snmp.engine_id", "")
		return nil, BadRequest(snmpv3UnsupportedError.Error())
	}

	bucket, ok := self.GetOrCreate(address, version_int, community)
	if ok {
		self.c <- bucket
	}

	bucket.byAddress.l.Lock()
	defer bucket.byAddress.l.Unlock()

	now := time.Now().Unix()
	bucket.updated_at = now

	list := bucket.snmpBuffer.All()
	if nil == list || 0 == len(list) {
		if !bucket.IsTimeout(now) {
			return nil, pendingError
		}
		return map[string]interface{}{"result": false, "list": list}, nil
	}

	return map[string]interface{}{"result": !bucket.IsTimeout(now), "list": list}, nil
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
