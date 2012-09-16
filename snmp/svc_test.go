package snmp

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type MockSvc struct {
	Svc
}

func (svc *MockSvc) Inc(a int) (resp int, err error) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err = NewPanicError("inc failed: ", recoverErr)
		}
	}()

	vals := svc.call(1*time.Second, svc.FuncOf(svc, "HandleInc"), a)
	resp = vals[0].(int)
	err = vals[1].(error)
	return
}

func (svc *MockSvc) HandleInc(a int) (resp int, err error) {
	resp = a
	resp++
	return
}

func (svc *MockSvc) Add(a, b int) (resp int, err error) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err = NewPanicError("add failed: ", recoverErr)
		}
	}()

	vals := svc.call(1*time.Second, svc.FuncOf(svc, "HandleAdd"), a, b)
	resp = vals[0].(int)
	err = vals[1].(error)
	return
}

func (svc *MockSvc) HandleAdd(a, b int) (resp int, err error) {
	resp = a + b
	return
}

func (svc *MockSvc) DivZero(a int) (resp int, err error) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err = NewPanicError("DivZero failed: ", recoverErr)
		}
	}()

	vals := svc.call(1*time.Second, svc.FuncOf(svc, "HandleDivZero"), a)
	resp = vals[0].(int)
	err = vals[1].(error)
	return
}

func (svc *MockSvc) HandleDivZero(a int) (resp int, err error) {
	err = errors.New("zero")
	resp = a + 10
	return
}

func (svc *MockSvc) throwPanic(a int) (resp int, err error) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err = NewPanicError("panic failed: ", recoverErr)
		}
	}()

	vals := svc.call(1*time.Second, svc.FuncOf(svc, "HandleThrowPanic"), a)
	resp = vals[0].(int)
	err = vals[1].(error)
	return
}

func (svc *MockSvc) HandleThrowPanic(a int) (resp int, err error) {
	err = errors.New("zero")
	resp = a + 10
	panic(errors.New("throwPanic"))
	return
}

func (svc *MockSvc) AddAsync(a int) (resp int, err error) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err = NewPanicError("panic failed: ", recoverErr)
		}
	}()

	vals := svc.call(3*time.Second, svc.FuncOf(svc, "HandleAddAsync"), a)
	resp = vals[0].(int)
	err = vals[1].(error)
	return
}

func (svc *MockSvc) HandleAddAsync(ctx InvokedContext, a int) (resp int, err error) {
	go func() {
		time.Sleep(2 * time.Second)
		ctx.Reply(10 + a)
	}()
	return
}

func (svc *MockSvc) SendInt(a int) (resp int, err error) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err = NewPanicError("panic failed: ", recoverErr)
		}
	}()

	var waiter sync.WaitGroup
	waiter.Add(1)

	svc.send(func() {
		resp = 20 + a
		waiter.Done()
	})

	timeouted := false
	go func() {
		time.Sleep(2 * time.Second)
		waiter.Done()
		timeouted = true
	}()
	waiter.Wait()
	if timeouted {
		err = errors.New("timeout")
	}
	return
}

func TestSvc(t *testing.T) {
	var mock MockSvc
	mock.Start()
	r, e := mock.Inc(3)
	if 4 != r {
		t.Errorf("inc error! %v", e)
	}

	r, e = mock.Add(5, 5)
	if 10 != r {
		t.Errorf("add error! %v", e)
	}
	r, e = mock.DivZero(4)
	if 14 != r || e.Error() != "zero" {
		t.Errorf("divzero error! %v", e)
	}
	_, e = mock.throwPanic(4)
	if !strings.Contains(e.Error(), "throwPanic") {
		t.Errorf("throwPanic error! %v", e.Error())
	}

	r, e = mock.AddAsync(5)
	if 15 != r {
		t.Errorf("add async error! %v", e)
	}

	r, e = mock.SendInt(5)
	if 25 != r && nil == e {
		t.Errorf("send int error! %v", e)
	}

	mock.Stop()
}
