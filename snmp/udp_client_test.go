package snmp

import (
	"encoding/hex"
	"fmt"
	"net"
	"testing"
)

func startServer() (net.PacketConn, net.Addr, error) {
	in, e := net.ListenPacket("udp", "127.0.0.1:0")
	if nil != e {
		return nil, nil, e
	}

	go serveTestUdp(in)
	return in, in.LocalAddr(), nil
}

func stopServer(in net.PacketConn) {
	in.Close()
}

func serveTestUdp(in net.PacketConn) {

	defer func() {
		fmt.Println("serveTestUdp exited!")
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

func TestClient(t *testing.T) {

	var listener net.PacketConn
	var cl Client

	defer func() {
		if nil != listener {
			fmt.Println("stopServer signal!")
			stopServer(listener)
		}
		if nil != cl && nil != cl.(Startable) {
			fmt.Println("stop client signal!")
			cl.(Startable).Stop()
		}
	}()

	listener, addr, err := startServer()
	if nil != err {
		t.Errorf("start udp server failed - %s", err.Error())
		return
	}

	fmt.Println(addr.String())
	cl, err = NewSnmpClient(addr.String())
	if nil != err {
		t.Errorf("create snmp client failed - %s", err.Error())
		return
	}
	if nil != cl.(Startable) {
		cl.(Startable).Start()
	}

	pdu, err := cl.CreatePDU(SNMP_PDU_GET, SNMP_V1)
	if nil != err {
		t.Errorf("create pdu failed - %s", err.Error())
		return
	}

	fmt.Println("begin send and recv!")
	res, err := cl.SendAndRecv(pdu)
	if nil != err {
		t.Errorf("sendAndRecv pdu failed - %s", err.Error())
		return
	}
	fmt.Println("end send and recv!")

	if nil == res {
		t.Errorf("sendAndRecv pdu failed - res is nil")
	}

	cl.FreePDU(pdu, res)
}
