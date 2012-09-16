package commons

import (
	"reflect"
	"time"
)

const (

)

type RpcResponse struct {
	messageType  int
	function     reflect.Value
	args         []reflect.Value
}

type RpcRequest struct {
	messageType  int
	args         []reflect.Value
}

type Rpc struct {
	ch chan RpcRequest
}

type Tctx struct {
	ch chan RpcResponse
}

//Calls a function with recover block
func (rpc *Rpc) safelyCall(function reflect.Value, args []reflect.Value) (resp []reflect.Value) {
	defer func() {
		if err := recover(); err != nil {
			rpc.panic_ch <- err
			resp = nil
		}
	}()

	return function.Call(args)
}

func RpcStart(svr interface{}) *Rpc {
	rpc := new(Rpc)
	rpc.send_ch = make(chan RpcMessage)
	return rpc
}

func (rpc *Rpc) Stop() {
	msg := &RpcMessage{nil, nil}
	rpc.send_ch <- msg
	select {
	case err := <-rpc.panic_ch:
		panic(err)
	case <-rpc.recv_ch:
		return
	case <-time.After(1 * time.Minute):
		panic(errors.New("timeout"))
	}
}

func (rpc *Rpc) Call(ch chan interface{}, function, args reflect.Value) []reflect.Value {
	msg := &RpcMessage{function, args}
	rpc.send_ch <- msg

	select {
	case chan := <-rpc.panic_ch:
		panic(err)
	case results := <-rpc.recv_ch:
		return results
	case <-time.After(1 * time.Minute):
		panic(errors.New("timeout"))
	}
}
