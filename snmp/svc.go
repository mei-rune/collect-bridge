package snmp

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PanicError struct {
	message string
	any     interface{}
}

func (err *PanicError) Error() string {
	return err.message
}

func NewPanicError(s string, any interface{}) error {
	return &PanicError{message: fmt.Sprintf("%s%v", s, any), any: any}
}

type TwinceError struct {
	message       string
	first, second error
}

func (err *TwinceError) Error() string {
	return err.message
}

func NewTwinceError(first, second error) error {
	msg := fmt.Sprintf("return two error, first is {%s}, second is {%s}",
		first.Error(), second.Error())
	return &TwinceError{message: msg, first: first, second: second}
}

const (
	MESSAGE_RET_PANIC = 0
	MESSAGE_RET_OK    = 1

	MESSAGE_TYPE_SYSTEM = 0
	MESSAGE_TYPE_USER   = 0

	MESSAGE_EXIT = 0
)

type Startable interface {
	Start() error
	Stop()
}

type InvokedContext interface {
	Reply(results ...interface{})
}

type responseMessage struct {
	responseType int
	panicResult  interface{}
	results      []interface{}
}

type requestMessage struct {
	messageType int
	function    *reflect.Value
	args        []reflect.Value
}

type message struct {
	ch       chan *message
	request  requestMessage
	response responseMessage

	isAsyncReply bool
}

func (msg *message) Reply(results ...interface{}) {
	msg.response.responseType = MESSAGE_RET_OK
	msg.response.results = results

	if nil != msg.ch {
		msg.ch <- msg
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////
// //var maxMessageCache = flag.Int("maxMessageCache", 1000, "set max size of response_message cache")
// var freeMessageList = make(chan *response_message, 1000)
//
// func getCachedMessage() (msg *response_message) {
//	select {
//	case msg = <-freeMessageList:
//	default:
//		msg = new(response_message)
//	}
//	return
// }
//
// func putCachedMessage(msg *response_message) {
//	select {
//	case freeMessageList <- msg:
//	default:
//	}
// }

//////////////////////////////////////////////////////////////////////////////////////////
//var maxChannelCache = flag.Int("maxChannelCache", 1000, "set max size of channel cache")
var freeChannelList = make(chan *message, 1000)

func getCachedChannel() (msg *message) {
	select {
	case msg = <-freeChannelList:
	default:
		msg = &message{ch: make(chan *message)}
	}
	msg.isAsyncReply = false
	return
}

func putCachedChannel(msg *message) {
	msg.request.args = nil
	msg.request.function = nil
	msg.response.results = nil
	msg.response.panicResult = nil
	msg.isAsyncReply = false
	select {
	case freeChannelList <- msg:
	default:
	}
}

const (
	status_inactive = 0
	status_active   = 1
)

type Svc struct {
	Name                    string
	Logger                  *log.Logger
	initialized             sync.Once
	status                  int32
	ch                      chan *message
	timeout                 time.Duration
	onStart, onStop, onIdle func()
}

func (svc *Svc) Set(onStart, onStop, onIdle func()) {
	svc.onStart = onStart
	svc.onStop = onStop
	svc.onIdle = onIdle
}

func (svc *Svc) FuncOf(target interface{}, name string) *reflect.Value {
	v := reflect.ValueOf(target)
	f := v.MethodByName(name)
	if 0 == f.Kind() {
		panic(fmt.Errorf("%s hasn`t method %s", v.String(), name))
	}
	return &f
}

func (svc *Svc) Start() (err error) {
	svc.initialized.Do(func() {
		svc.status = status_inactive
		svc.ch = make(chan *message)
		if "" == svc.Name {
			svc.Name = "SVC-" + strconv.Itoa(time.Now().Second())
		}

		if nil == svc.Logger {
			svc.Logger = log.New(os.Stdout, svc.Name+" - ", log.LstdFlags|log.Lshortfile)
		} else if "" == svc.Logger.Prefix() {
			svc.Logger.SetPrefix(svc.Name + " - ")
			svc.Logger.SetFlags(svc.Logger.Flags() | log.Lshortfile)
		}
	})
	if !atomic.CompareAndSwapInt32(&svc.status, status_inactive, status_active) {
		return
	}

	go serve(svc)

	msg := <-svc.ch
	if MESSAGE_RET_PANIC == msg.response.responseType {
		err = msg.response.panicResult.(error)
		return
	}

	return
}

func (svc *Svc) Stop() {
	if !atomic.CompareAndSwapInt32(&svc.status, status_active, status_inactive) {
		svc.Logger.Printf("It is already exited\r\n")
		return
	}

	msg := getCachedChannel()
	var success bool = false
	defer func() {
		if success {
			putCachedChannel(msg)
		}
	}()

	msg.request.messageType = MESSAGE_TYPE_SYSTEM
	msg.request.function = nil
	svc.ch <- msg
	select {
	case <-msg.ch:
		success = true
	case <-time.After(5 * time.Minute):
		panic(errors.New("time out"))
	}
	return
}

func function2Value(function interface{}) (fvp *reflect.Value) {
	var ok bool
	var fv reflect.Value

	fvp, ok = function.(*reflect.Value)
	if ok {
		goto end
	}

	fv, ok = function.(reflect.Value)
	if !ok {
		fv = reflect.ValueOf(function)
	}

	fvp = &fv

end:
	if reflect.Func != fvp.Kind() {
		panic(fmt.Errorf("paramter 'function' isn't a function. real type is %v", fvp.Kind()))
	}
	return
}

func (svc *Svc) Send(function interface{}, args ...interface{}) {
	if !svc.IsAlive() {
		panic(errors.New("svc is stopped."))
		return
	}

	msg := new(message)

	msg.request.messageType = MESSAGE_TYPE_USER
	msg.request.function = function2Value(function)
	msg.request.args = make([]reflect.Value, len(args))
	for i, v := range args {
		msg.request.args[i] = reflect.ValueOf(v)
	}

	svc.ch <- msg
	return
}

func panicArgumentsError(in, n int) {
	if in < n {
		panic(errors.New("call: Call with too few input arguments"))
	}
	panic(errors.New("call: Call with too many input arguments"))
}

func (svc *Svc) SafelyCall(timeout time.Duration, function interface{}, args ...interface{}) (results []interface{}) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err := NewPanicError("SafelyCall failed: ", recoverErr)
			switch {
			case nil == results || 0 == len(results):
				results = []interface{}{err}
			case nil != results[len(results)].(error):
				results[len(results)-1] = NewTwinceError(results[len(results)-1].(error), err)
			default:
				results = append(results, err)
			}
		}
	}()

	results = svc.innerCall(timeout, function, args)
	return
}

func (svc *Svc) Call(timeout time.Duration, function interface{}, args ...interface{}) []interface{} {
	return svc.innerCall(timeout, function, args)
}

func (svc *Svc) innerCall(timeout time.Duration, function interface{}, args []interface{}) (results []interface{}) {
	if !svc.IsAlive() {
		panic(errors.New("svc is stopped."))
		return
	}

	msg := getCachedChannel()
	var success bool = false

	defer func() {
		if success {
			putCachedChannel(msg)
		} else {
			if nil != msg.ch {
				close(msg.ch)
			}
		}
	}()

	msg.request.messageType = MESSAGE_TYPE_USER
	msg.request.function = function2Value(function)

	var argsLen int = len(args)
	funcType := msg.request.function.Type()
	if inLen := funcType.NumIn(); inLen != argsLen {

		if inLen != (argsLen + 1) {
			panicArgumentsError(inLen, argsLen)
		}

		msgType := reflect.TypeOf(msg)
		if !msgType.AssignableTo(funcType.In(0)) {
			panicArgumentsError(inLen, argsLen)
		}

		msg.request.args = make([]reflect.Value, argsLen+1)
		msg.request.args[0] = reflect.ValueOf(msg)
		msg.isAsyncReply = true

		for i, v := range args {
			msg.request.args[i+1] = reflect.ValueOf(v)
		}
	} else {
		msg.request.args = make([]reflect.Value, argsLen)
		msg.isAsyncReply = false

		for i, v := range args {
			msg.request.args[i] = reflect.ValueOf(v)
		}
	}

	svc.ch <- msg
	select {
	case resp := <-msg.ch:
		switch resp.response.responseType {
		case MESSAGE_RET_PANIC:
			success = true
			panic(resp.response.panicResult)
		case MESSAGE_RET_OK:
			success = true
			results = resp.response.results
		default:
			panic(fmt.Errorf("unknown response type:", resp.response.responseType))
		}
	case <-time.After(timeout):
		panic(errors.New("time out"))
	}
	return
}

func serve(svc *Svc) {

	var exit_ch chan *message = nil
	var exit_msg *message = nil
	isStarted := false

	defer func() {
		atomic.CompareAndSwapInt32(&svc.status, status_active, status_inactive)
		if isStarted {
			handleExit(svc)
		}
		if err := recover(); nil != err {
			svc.Logger.Printf("%v\r\n", err)
		}
		if nil != exit_ch {
			exit_ch <- exit_msg
		}
	}()

	if 0 == svc.timeout {
		svc.timeout = 10 * time.Second
	}

	if nil == svc.ch {
		svc.Logger.Println("start svc failed, ch is nil!")
		return
	}

	start_resp_msg := &message{}

	se := handleStart(svc)
	if nil != se {
		start_resp_msg.response.responseType = MESSAGE_RET_PANIC
		start_resp_msg.response.panicResult = se
		exit_msg = start_resp_msg
		exit_ch = svc.ch
		return
	}
	start_resp_msg.response.responseType = MESSAGE_RET_OK
	svc.ch <- start_resp_msg
	isStarted = true

	for {
		select {
		case msg, ok := <-svc.ch:
			if !ok {
				goto exit
			}
			if msg.request.messageType == MESSAGE_TYPE_SYSTEM {
				if nil == msg.request.function {
					if nil != msg.ch {
						exit_msg = msg
						exit_ch = msg.ch
					}
					goto exit
				}
			}
			svc.safelyCall(msg)
		case <-time.After(svc.timeout):
			handleTick(svc)
		}
	}
exit:
	svc.Logger.Printf("channel is closed or recv an exit message!\r\n")
}

func handleStart(svc *Svc) (err error) {

	defer func() {
		if e := recover(); nil != e {
			err = NewPanicError("call onStart failed - ", e)
		}
	}()

	if nil != svc.onStart {
		svc.onStart()
	}
	return
}

func handleExit(svc *Svc) {

	defer func() {
		if err := recover(); nil != err {
			svc.Logger.Printf("on exit - %v\r\n", err)
		}
	}()

	if nil != svc.onStop {
		svc.onStop()
	}
}

func handleTick(svc *Svc) {

	defer func() {
		if err := recover(); nil != err {
			svc.Logger.Printf("on idle - %v\r\n", err)
		}
	}()

	if nil != svc.onIdle {
		svc.onIdle()
	}
}

func (svc *Svc) IsAlive() bool {
	return status_active == atomic.LoadInt32(&svc.status)
}

func (svc *Svc) safelyCall(msg *message) {
	if nil == msg {
		return
	}

	defer func() {
		if err := recover(); nil != err {
			svc.Logger.Printf("%v\r\n", err)
		}
	}()

	ret := svc.callMessage(msg)
	if nil != msg.ch && !msg.isAsyncReply {
		msg.ch <- ret
	}
}

func (svc *Svc) callMessage(msg *message) (reply *message) {
	reply = msg
	reply.response.responseType = MESSAGE_RET_OK
	defer func() {
		if err := recover(); nil != err {
			reply.response.responseType = MESSAGE_RET_PANIC
			reply.response.panicResult = err
		}
	}()

	results := msg.request.function.Call(msg.request.args)
	if !msg.isAsyncReply {
		reply.response.results = make([]interface{}, len(results))
		for i, v := range results {
			reply.response.results[i] = v.Interface()
		}
	}
	return
}
