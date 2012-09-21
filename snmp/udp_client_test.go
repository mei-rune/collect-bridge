package snmp

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func startServer(laddr string) (net.PacketConn, net.Addr, *sync.WaitGroup, error) {
	in, e := net.ListenPacket("udp", laddr)
	if nil != e {
		return nil, nil, nil, e
	}

	var waiter sync.WaitGroup
	waiter.Add(1)

	go serveTestUdp(in, &waiter)

	return in, in.LocalAddr(), &waiter, nil
}

func stopServer(in net.PacketConn) {
	in.Close()
}

func serveTestUdp(in net.PacketConn, waiter *sync.WaitGroup) {

	defer func() {
		waiter.Done()
	}()

	var bytes [10000]byte

	for {
		_, addr, err := in.ReadFrom(bytes[:])
		if nil != err {
			fmt.Println(err.Error())
			break
		}

		bin, err := hex.DecodeString(snmpv1_txt)
		if nil != err {
			fmt.Println(err.Error())
		} else {
			in.WriteTo(bin, addr)
		}
	}
}

type callback func(t *testing.T, cl Client, laddr net.Addr)

func TestReturnPdu(t *testing.T) {
	testWith(t, "127.0.0.1:0", func(t *testing.T, cl Client, laddr net.Addr) {

		cl.(*UdpClient).requestId = 233
		pdu, err := cl.CreatePDU(SNMP_PDU_GET, SNMP_V1)
		if nil != err {
			t.Errorf("create pdu failed - %s", err.Error())
			return
		}

		res, err := cl.SendAndRecv(pdu, 2*time.Second)
		if nil != err {
			t.Errorf("sendAndRecv pdu failed - %s", err.Error())
			return
		}

		if nil == res {
			t.Errorf("sendAndRecv pdu failed - res is nil")
		}

		cl.FreePDU(pdu, res)
	})
}

func TestSendFailed(t *testing.T) {
	testWith(t, "0.0.0.0:0", func(t *testing.T, cl Client, laddr net.Addr) {

		cl.(*UdpClient).requestId = 233
		pdu, err := cl.CreatePDU(SNMP_PDU_GET, SNMP_V1)
		if nil != err {
			t.Errorf("create pdu failed - %s", err.Error())
			return
		}

		_, err = cl.SendAndRecv(pdu, 2*time.Second)
		if nil == err {
			t.Errorf("excepted throw an error, actual return ok")
			return
		}
		cl.FreePDU(pdu)
	})
}

func TestRecvTimeout(t *testing.T) {
	testWith(t, "127.0.0.1:0", func(t *testing.T, cl Client, laddr net.Addr) {
		pdu, err := cl.CreatePDU(SNMP_PDU_GET, SNMP_V1)
		if nil != err {
			t.Errorf("create pdu failed - %s", err.Error())
			return
		}

		_, err = cl.SendAndRecv(pdu, 2*time.Second)
		if nil == err {
			t.Errorf("excepted throw an error, actual return ok")
			return
		}
		fmt.Println("----------")
		fmt.Println(err)
		cl.FreePDU(pdu)
	})
}

func testWith(t *testing.T, laddr string, f callback) {
	var waiter *sync.WaitGroup
	var listener net.PacketConn
	var cl Client

	defer func() {
		if nil != listener {
			stopServer(listener)
			waiter.Wait()
		}
		if nil != cl && nil != cl.(Startable) {
			cl.(Startable).Stop()
		}
	}()

	listener, addr, waiter, err := startServer(laddr)
	if nil != err {
		t.Errorf("start udp server failed - %s", err.Error())
		return
	}

	cl, err = NewSnmpClient(addr.String())
	if nil != err {
		t.Errorf("create snmp client failed - %s", err.Error())
		return
	}
	if nil != cl.(Startable) {
		cl.(Startable).Start()
	}

	f(t, cl, addr)

}
