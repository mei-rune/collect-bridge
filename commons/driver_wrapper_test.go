package commons

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type mock_driver struct {
	started, stopped        bool
	start_panic, stop_panic string
	start_error             error
}

func (svc *mock_driver) Start() error {
	svc.started = true
	if "" != svc.start_panic {
		panic(svc.start_panic)
	}

	return svc.start_error
}

func (svc *mock_driver) Stop() {
	svc.stopped = true
	if "" != svc.stop_panic {
		panic(svc.stop_panic)
	}
}

func (svc *mock_driver) Get(params map[string]string) Result {
	resp, _ := strconv.Atoi(params["a"])
	resp++
	return Return(nil).SetOption("result", resp)
}

func (svc *mock_driver) Put(params map[string]string) Result {
	a, _ := strconv.Atoi(params["a"])
	b, _ := strconv.Atoi(params["b"])

	return Return(nil).SetOption("result", a+b)
}

func (svc *mock_driver) Create(params map[string]string) Result {
	a, _ := strconv.Atoi(params["a"])
	b, _ := strconv.Atoi(params["b"])

	return Return(nil).SetOption("result", a-b)
}

func (svc *mock_driver) Delete(params map[string]string) Result {
	a, _ := params["a"]
	b, _ := params["b"]

	return Return(nil).SetOption("result", a+b).SetError(500, a+b)
}

func TestDriverWrapperStartFailed(t *testing.T) {
	mock := Wrap(&mock_driver{start_error: errors.New("this is error")}, 5*time.Second)
	if nil != mock.Start() {
		t.Errorf("start svc failed, expect return a error, actual is success.")
	}

	mock = Wrap(&mock_driver{start_panic: "this is error"}, 5*time.Second)
	if nil != mock.Start() {
		t.Errorf("start svc failed, expect return a error, actual is success.")
	}
}

func TestDriverWrapperStartedAndStopped(t *testing.T) {
	mock := &mock_driver{}
	wrap := Wrap(mock, 5*time.Second)
	wrap.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	wrap.Start()
	wrap.Stop()

	if !mock.started {
		t.Error("start error")
	}

	if !mock.stopped {
		t.Error("stop error")
	}

	mock = &mock_driver{stop_panic: "throw a error"}
	wrap = Wrap(mock, 5*time.Second)
	wrap.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	wrap.Start()
	wrap.Stop()

	if !mock.started {
		t.Error("start error")
	}

	if !mock.stopped {
		t.Error("stop error")
	}
}

func TestDriverWrapper(t *testing.T) {
	mock := Wrap(&mock_driver{}, 5*time.Second)
	mock.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)

	mock.Start()
	defer mock.Stop()

	result := mock.Get(map[string]string{"a": "1"})
	if result.HasError() {
		t.Errorf("get error! %v", result.ErrorMessage())
	}
	if !reflect.DeepEqual(Return(nil).SetOption("result", 2), result) {
		t.Errorf("get error, excepted is %v, actual is %v", result, Return(nil).SetOption("result", 2))
	}
	result = mock.Put(map[string]string{"a": "1", "b": "3"})
	if result.HasError() {
		t.Errorf("put error! %v", result.ErrorMessage())
	}
	if !reflect.DeepEqual(Return(nil).SetOption("result", 4), result) {
		t.Errorf("put error, excepted is %v, actual is %v", result, Return(nil).SetOption("result", 4))
	}
	result = mock.Create(map[string]string{"a": "9", "b": "3"})
	if result.HasError() {
		t.Errorf("create error! %v", result.ErrorMessage())
	}
	if !reflect.DeepEqual(Return(nil).SetOption("result", 6), result) {
		t.Errorf("create error, excepted is %v, actual is %v", result, Return(nil).SetOption("result", 6))
	}
	result = mock.Delete(map[string]string{"a": "9", "b": "3"})
	if !result.HasError() {
		t.Errorf("delete error! %v", result.ErrorMessage())
	} else {
		if result.ErrorMessage() != "93" {
			t.Errorf("create error, excepted is %v, actual is %v", result.ErrorMessage(), "93")
		}
	}

}
