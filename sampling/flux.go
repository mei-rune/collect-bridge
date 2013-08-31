package sampling

import (
	"commons"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Flux struct {
	IfInOctets        uint64 `json:"IfInOctets"`
	IfInUcastPkts     uint64 `json:"IfInUcastPkts"`
	IfInNUcastPkts    uint64 `json:"IfInNUcastPkts"`
	IfInDiscards      uint64 `json:"IfInDiscards"`
	IfInErrors        uint64 `json:"IfInErrors"`
	IfInUnknownProtos uint64 `json:"IfInUnknownProtos"`
	IfOutOctets       uint64 `json:"IfOutOctets"`
	IfOutUcastPkts    uint64 `json:"IfOutUcastPkts"`
	IfOutNUcastPkts   uint64 `json:"IfOutNUcastPkts"`
	IfOutDiscards     uint64 `json:"IfOutDiscards"`
	IfOutErrors       uint64 `json:"IfOutErrors"`
	SampledAt         int64  `json:"sampled_at"`
	IfIndex           int    `json:"IfIndex"`
	IfBit             int
	IfStatus          int `json:"IfStatus"`
}

func uint64With(row map[string]interface{}, key string) uint64 {
	v, ok := row[key]
	if !ok {
		panic("'" + key + "' is not exists.")
	}
	switch value := v.(type) {
	case uint64:
		return value
	case uint32:
		return uint64(value)
	case uint:
		return uint64(value)
	default:
		panic(fmt.Sprintf("'%v' is not a uint32, uint64 or uint - %T.", v, v))
	}
}

func intWith(row map[string]interface{}, key string) int {
	v, ok := row[key]
	if !ok {
		panic("'" + key + "' is not exists.")
	}
	switch value := v.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case int32:
		return int(value)
	default:
		panic(fmt.Sprintf("'%v' is not a int32, int64 or int - %T.", v, v))
	}
}

func readFluxFormMap(flux *Flux, new_row map[string]interface{}) {
	flux.IfBit = intWith(new_row, "ifBit")
	flux.IfStatus = intWith(new_row, "ifStatus")
	flux.IfInOctets = uint64With(new_row, "ifInOctets")
	flux.IfInUcastPkts = uint64With(new_row, "ifInUcastPkts")
	flux.IfInNUcastPkts = uint64With(new_row, "ifInNUcastPkts")
	flux.IfInDiscards = uint64With(new_row, "ifInDiscards")
	flux.IfInErrors = uint64With(new_row, "ifInErrors")
	flux.IfInUnknownProtos = uint64With(new_row, "ifInUnknownProtos")
	flux.IfOutOctets = uint64With(new_row, "ifOutOctets")
	flux.IfOutUcastPkts = uint64With(new_row, "ifOutUcastPkts")
	flux.IfOutNUcastPkts = uint64With(new_row, "ifOutNUcastPkts")
	flux.IfOutDiscards = uint64With(new_row, "ifOutDiscards")
	flux.IfOutErrors = uint64With(new_row, "ifOutErrors")
}

const MAX_UINT64_VALUE uint64 = 18446744073709551615
const MAX_UINT32_VALUE uint64 = 4294967295
const DEFAULT_REVERSAL uint64 = 3000

func isReversal(ifLast, ifCurrent uint64) bool {
	return ifLast-ifCurrent > DEFAULT_REVERSAL
}

type calcFunc func(ifLast, ifCurrent uint64) uint64

func calc32Bit(ifLast, ifCurrent uint64) uint64 {
	if isReversal(ifLast, ifCurrent) {
		return (MAX_UINT32_VALUE - ifLast) + ifCurrent
	} else if ifCurrent < ifLast {
		return ifLast - ifCurrent
	} else {
		return ifCurrent - ifLast
	}
}

func calc64Bit(ifLast, ifCurrent uint64) uint64 {
	if isReversal(ifLast, ifCurrent) {
		return (MAX_UINT64_VALUE - ifLast) + ifCurrent
	} else if ifCurrent < ifLast {
		return ifLast - ifCurrent
	} else {
		return ifCurrent - ifLast
	}
}

var pendingError = errors.New("sample is pending.")

func calcFlux(res *Flux, buffer *fluxBuffer, interval uint64, last_at int64) error {
	if 2 > buffer.Size() {
		return pendingError
	}

	current := buffer.Last()
	if 1 != current.IfStatus {
		res.IfBit = 0
		res.IfStatus = current.IfStatus
		res.IfInOctets = 0
		res.IfInUcastPkts = 0
		res.IfInNUcastPkts = 0
		res.IfInDiscards = 0
		res.IfInErrors = 0
		res.IfInUnknownProtos = 0
		res.IfOutOctets = 0
		res.IfOutUcastPkts = 0
		res.IfOutNUcastPkts = 0
		res.IfOutDiscards = 0
		res.IfOutErrors = 0
		return nil
	}

	time_interval := uint64(0)

	var calc calcFunc
	switch current.IfBit {
	case 32:
		calc = calc32Bit
	case 64:
		calc = calc64Bit
	default:
		return errors.New("unsupported bitSize - " + strconv.FormatInt(int64(current.IfBit), 10))
	}

	var e error
	for i := buffer.Size() - 2; i >= 0; i-- {
		last := buffer.Get(i)
		if nil == last {
			break
		}
		if last.IfBit != current.IfBit {
			e = errors.New("IfBit of current is not same with last")
			break
		}
		if last.IfStatus != current.IfStatus {
			e = errors.New("IfStatus of current is not same with last")
			break
		}
		if current.SampledAt < last.SampledAt {
			e = fmt.Errorf("sampledAt(%v) of current is less whit last(%v)", current.SampledAt, last.SampledAt)
			break
		}

		res.IfInOctets += calc(last.IfInOctets, current.IfInOctets)
		res.IfInUcastPkts += calc(last.IfInUcastPkts, current.IfInUcastPkts)
		res.IfInNUcastPkts += calc(last.IfInNUcastPkts, current.IfInNUcastPkts)
		res.IfInDiscards += calc(last.IfInDiscards, current.IfInDiscards)
		res.IfInErrors += calc(last.IfInErrors, current.IfInErrors)
		res.IfInUnknownProtos += calc(last.IfInUnknownProtos, current.IfInUnknownProtos)
		res.IfOutOctets += calc(last.IfOutOctets, current.IfOutOctets)
		res.IfOutUcastPkts += calc(last.IfOutUcastPkts, current.IfOutUcastPkts)
		res.IfOutNUcastPkts += calc(last.IfOutNUcastPkts, current.IfOutNUcastPkts)
		res.IfOutDiscards += calc(last.IfOutDiscards, current.IfOutDiscards)
		res.IfOutErrors += calc(last.IfOutErrors, current.IfOutErrors)
		if last_at >= last.SampledAt {
			break
		}
		time_interval += uint64(current.SampledAt - last.SampledAt)
		if time_interval > interval {
			break
		}

		current = last
	}

	if 0 == time_interval {
		if nil == e {
			return pendingError
		}
		return e
	}

	res.IfInOctets = res.IfInOctets / time_interval
	res.IfInUcastPkts = res.IfInUcastPkts / time_interval
	res.IfInNUcastPkts = res.IfInNUcastPkts / time_interval
	res.IfInDiscards = (res.IfInDiscards * 60) / time_interval
	res.IfInErrors = (res.IfInErrors * 60) / time_interval
	res.IfInUnknownProtos = (res.IfInUnknownProtos * 60) / time_interval
	res.IfOutOctets = res.IfOutOctets / time_interval
	res.IfOutUcastPkts = res.IfOutUcastPkts / time_interval
	res.IfOutNUcastPkts = res.IfOutNUcastPkts / time_interval
	res.IfOutDiscards = (res.IfOutDiscards * 60) / time_interval
	res.IfOutErrors = (res.IfOutErrors * 60) / time_interval
	return nil
}

type linkBucket struct {
	l sync.Mutex
	fluxBuffer
}

type linkWorker struct {
	get         portScalar
	l           sync.Mutex
	fluxBuffers map[int]*linkBucket
}

func (self *linkWorker) OnTick() {
}

func (self *linkWorker) Get(id int) *linkBucket {
	self.l.Lock()
	defer self.l.Unlock()

	if buffer, ok := self.fluxBuffers[id]; ok {
		return buffer
	}

	w := &linkBucket{}
	w.fluxBuffer.init(make([]Flux, *flux_buffer_size))
	self.fluxBuffers[id] = w
	return w
}

func (self *linkWorker) Close() {
}

func (self *linkWorker) Stats() map[string]interface{} {
	self.l.Lock()
	defer self.l.Unlock()

	stats := make([]int, 0, len(self.fluxBuffers))
	for k, _ := range self.fluxBuffers {
		stats = append(stats, k)
	}

	return map[string]interface{}{"name": "linkWorker", "ports": stats}
}

func (self *linkWorker) Call(ctx MContext) commons.Result {
	id, e := ctx.GetInt("id")
	if nil != e {
		return commons.ReturnWithIsRequired("id")
	}

	// forward, e := ctx.GetBool("@forward")
	// if nil != e {
	// 	return commons.ReturnWithIsRequired("forward")
	// }
	from_based, e := ctx.GetBool("@from_based")
	if nil != e {
		return commons.ReturnWithIsRequired("from_based")
	}
	var ifIndex string
	var device string

	if from_based {
		device, e = ctx.GetString("@from_device")
		if nil != e {
			return commons.ReturnWithIsRequired("from_device")
		}
		ifIndex, e = ctx.GetString("@from_if_index")
		if nil != e {
			return commons.ReturnWithIsRequired("from_if_index")
		}
	} else {
		device, e = ctx.GetString("@to_device")
		if nil != e {
			return commons.ReturnWithIsRequired("to_device")
		}
		ifIndex, e = ctx.GetString("@to_if_index")
		if nil != e {
			return commons.ReturnWithIsRequired("to_if_index")
		}
	}

	max_interval := ctx.GetUint64WithDefault("max_interval", 0)
	last_at := ctx.GetInt64WithDefault("last_at", 0)
	if 0 == last_at && 0 == max_interval {
		max_interval = 20
	}

	bucket := self.Get(id)
	bucket.l.Lock()
	defer bucket.l.Unlock()

	ctx2, e := ctx.Read().CreateCtx("interface", "managed_object", device)
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}
	attributes, e := ctx.Read().GetObject("interface", []P{P{"port", ifIndex}}, ctx2)
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}
	sampled_at := time.Now().Unix()
	flux := bucket.BeginPush()
	readFluxFormMap(flux, attributes)
	flux.SampledAt = sampled_at
	bucket.CommitPush()

	var current Flux
	if e := calcFlux(&current, &bucket.fluxBuffer, max_interval, last_at); nil != e {
		fmt.Println(bucket.fluxBuffer.All())
		return commons.ReturnWithInternalError(e.Error())
	}
	current.SampledAt = sampled_at
	return commons.Return(&current)
}

type interfaceBucket struct {
	l sync.Mutex
	fluxBuffer
}

type interfaceWorker struct {
	get         portScalar
	l           sync.Mutex
	fluxBuffers map[string]*interfaceBucket
}

func (self *interfaceWorker) OnTick() {
}

func (self *interfaceWorker) Get(id string) *interfaceBucket {
	self.l.Lock()
	defer self.l.Unlock()

	if buffer, ok := self.fluxBuffers[id]; ok {
		return buffer
	}

	w := &interfaceBucket{}
	w.fluxBuffer.init(make([]Flux, *flux_buffer_size))
	self.fluxBuffers[id] = w
	return w
}

func (self *interfaceWorker) Close() {
}

func (self *interfaceWorker) Stats() map[string]interface{} {
	self.l.Lock()
	defer self.l.Unlock()

	stats := make([]string, 0, len(self.fluxBuffers))
	for k, _ := range self.fluxBuffers {
		stats = append(stats, k)
	}

	return map[string]interface{}{"name": "interfaceWorker", "ports": stats}
}

func (self *interfaceWorker) Call(ctx MContext) commons.Result {
	ifIndex, e := ctx.GetString("@ifIndex")
	if nil != e || 0 == len(ifIndex) {
		return commons.ReturnWithIsRequired("ifIndex")
	}

	ifIndex_int, e := strconv.ParseInt(ifIndex, 10, 0)
	if nil != e {
		return commons.ReturnWithBadRequest("'ifIndex' is not a int.")
	}

	uid, e := ctx.GetString("uid")
	if nil != e || 0 == len(ifIndex) {
		return commons.ReturnWithIsRequired("uid")
	}

	max_interval := ctx.GetUint64WithDefault("max_interval", 0)
	last_at := ctx.GetInt64WithDefault("last_at", 0)
	if 0 == last_at && 0 == max_interval {
		max_interval = 20
	}

	bucket := self.Get(uid + "/" + ifIndex)
	bucket.l.Lock()
	defer bucket.l.Unlock()
	attributes, e := ctx.Read().GetObject("interface", []P{P{"port", ifIndex}}, ctx)
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}
	sampled_at := time.Now().Unix()
	flux := bucket.BeginPush()
	readFluxFormMap(flux, attributes)
	flux.IfIndex = int(ifIndex_int)
	flux.SampledAt = sampled_at
	bucket.CommitPush()

	var current Flux
	if e := calcFlux(&current, &bucket.fluxBuffer, max_interval, last_at); nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}
	current.SampledAt = sampled_at

	return commons.Return(&current)
}

func init() {

	Methods["default_interface_flux"] = newRouteSpec("get", "interface_flux", "interface_flux", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			v, ok := params["backgroundWorkers"]
			if !ok {
				return nil, commons.IsRequired("backgroundWorkers")
			}

			bw, ok := v.(BackgroundWorkers)
			if !ok {
				return nil, commons.TypeError("'backgroundWorkers' isn't a BackgroundWorkers")
			}

			drv := &interfaceWorker{fluxBuffers: make(map[string]*interfaceBucket)}
			e := drv.get.Init(params)
			if nil != e {
				return nil, e
			}

			bw.Add("interfaceWorker", drv)
			return drv, nil
		})

	Methods["default_interface_flux_native"] = newRouteSpecWithPaths("get", "interface_flux", "interface_flux", []P{P{"port", "@ifIndex"}}, nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			v, ok := params["backgroundWorkers"]
			if !ok {
				return nil, commons.IsRequired("backgroundWorkers")
			}

			bw, ok := v.(BackgroundWorkers)
			if !ok {
				return nil, commons.TypeError("'backgroundWorkers' isn't a BackgroundWorkers")
			}

			drv := &interfaceWorker{fluxBuffers: make(map[string]*interfaceBucket)}
			e := drv.get.Init(params)
			if nil != e {
				return nil, e
			}

			bw.Add("interfaceWorker", drv)
			return drv, nil
		})

	Methods["default_link_flux"] = newRouteSpec("get", "link_flux", "link_flux", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			v, ok := params["backgroundWorkers"]
			if !ok {
				return nil, commons.IsRequired("backgroundWorkers")
			}

			bw, ok := v.(BackgroundWorkers)
			if !ok {
				return nil, commons.TypeError("'backgroundWorkers' isn't a BackgroundWorkers")
			}

			drv := &linkWorker{fluxBuffers: make(map[int]*linkBucket)}
			e := drv.get.Init(params)
			if nil != e {
				return nil, e
			}

			bw.Add("linkWorker", drv)
			return drv, nil
		})
}
