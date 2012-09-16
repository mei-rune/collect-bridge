package snmp

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"time"
)

type Request struct {
	pdu PDU
	ctx InvokedContext
}

type UdpClient struct {
	host     string
	engine   snmpEngine
	conn     *net.UDPConn
	pendings map[int]*Request
	Svc
}

func NewSnmpClient(host string) (Client, error) {
	cl := &UdpClient{host: host}
	cl.pendings = make(map[int]*Request)
	return cl, nil
}

func (client *UdpClient) CreatePDU(op, version int) (PDU, error) {
	if op < 0 || SNMP_PDU_REPORT < op {
		return nil, errors.New(fmt.Sprintf("unsupported pdu type: %d", op))
	}

	switch version {
	case SNMP_V1, SNMP_V2C:
		return &V2CPDU{op: op, client: client, version: version}, nil
	case SNMP_V3:
		return &V3PDU{op: op, client: client}, nil
	}
	return nil, errors.New(fmt.Sprintf("unsupported version: %d", version))
}

func (client *UdpClient) SendAndRecv(req PDU) (pdu PDU, err error) {

	defer func() {
		if recoverErr := recover(); nil != recoverErr {
			err = NewPanicError("send/recv a pdu failed: ", recoverErr)
		}
	}()

	values := client.call(5*time.Minute, func(ctx InvokedContext) { client.handleSend(ctx, req) })
	if nil != values[0] {
		pdu = values[0].(PDU)
	}
	if nil != values[1] {
		err = values[1].(error)
	}
	return
}

func (client *UdpClient) createConnect() (err error) {
	if nil != client.conn {
		return nil
	}
	addr, err := net.ResolveUDPAddr("udp", client.host)
	if nil != err {
		return err
	}
	client.conn, err = net.DialUDP("udp", nil, addr)
	if nil != err {
		return err
	}

	go client.readUDP(client.conn)
	return nil
}

var maxPDUSize = flag.Int("maxPDUSize", 10240, "set max size of pdu")

func (client *UdpClient) readUDP(conn *net.UDPConn) {
	defer func() {
		if err := recover(); nil != err {
			fmt.Printf("%v\r\n", err)
		}
		conn.Close()
	}()

	for {
		bytes := make([]byte, *maxPDUSize)
		length, err := conn.Read(bytes)
		if nil != err {
			fmt.Printf("%v\r\n", err)
			break
		}

		func(buf []byte) {
			client.send(func() { client.handleRecv(buf) })
		}(bytes[:length])
	}
}

func (client *UdpClient) handleRecv(bytes []byte) {
	pdu, err := DecodePDU(bytes)
	if nil != err {
		fmt.Printf("%v\r\n", err)
		return
	}

	req, ok := client.pendings[pdu.GetRequestID()]
	if !ok {
		fmt.Println("not found request.")
		return
	}

	req.ctx.Reply(pdu, nil)
}

func (client *UdpClient) handleSend(ctx InvokedContext, pdu PDU) {
	var err error = nil
	var ok bool = false
	var bytes []byte = nil
	if nil == client.conn {
		err = client.createConnect()
		if nil != err {
			goto failed
		}
	}

	_, ok = client.pendings[pdu.GetRequestID()]
	if ok {
		err = errors.New("identifier is repected.")
		goto failed
	}

	bytes, err = EncodePDU(pdu)
	if nil != err {
		err = errors.New("encode pdu failed - " + err.Error())
		goto failed
	}

	_, err = client.conn.Write(bytes)
	if nil != err {
		err = errors.New("send pdu failed - " + err.Error())
		goto failed
	}

	client.pendings[pdu.GetRequestID()] = &Request{pdu: pdu, ctx: ctx}
	return
failed:
	ctx.Reply(nil, err)
	return
}

func (client *UdpClient) FreePDU(pdus ...PDU) {

}
