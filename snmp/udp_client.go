package snmp

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

var maxPDUSize = flag.Int("maxPDUSize", 10240, "set max size of pdu")
var logPdu = flag.Bool("log.pdu", false, "log pdu content?")

type Request struct {
	pdu PDU
	ctx InvokedContext
}

type UdpClient struct {
	requestId int
	host      string
	engine    snmpEngine
	conn      *net.UDPConn
	pendings  map[int]*Request
	Svc
}

func NewSnmpClient(host string) (Client, error) {
	cl := &UdpClient{host: NormalizeAddress(host)}
	cl.pendings = make(map[int]*Request)
	return cl, nil
}

func (client *UdpClient) CreatePDU(op SnmpType, version SnmpVersion) (PDU, error) {
	if op < 0 || SNMP_PDU_REPORT < op {
		return nil, fmt.Errorf("unsupported pdu type: %d", op)
	}

	switch version {
	case SNMP_V1, SNMP_V2C:
		return &V2CPDU{op: op, client: client, version: version}, nil
	case SNMP_V3:
		return &V3PDU{op: op, client: client}, nil
	}
	return nil, fmt.Errorf("unsupported version: %d", version)
}

func handleRemoveRequest(client *UdpClient, id int) {
	delete(client.pendings, id)
}

func (client *UdpClient) SendAndRecv(req PDU, timeout time.Duration) (pdu PDU, err error) {

	defer func() {
		if 0 != req.GetRequestID() {
			client.Send(handleRemoveRequest, client, req.GetRequestID())
		}
	}()

	values := client.SafelyCall(timeout, func(ctx InvokedContext) { client.handleSend(ctx, req) })
	switch len(values) {
	case 0:
		err = fmt.Errorf("return empty.")
	case 1:
		if nil != values[0] {
			err = values[0].(error)
		}
	case 2:
		if nil != values[0] {
			pdu = values[0].(PDU)
		}
		if nil != values[1] {
			err = values[1].(error)
		}
	default:
		err = fmt.Errorf("num of return value is error.")
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

func (client *UdpClient) readUDP(conn *net.UDPConn) {
	defer func() {
		if err := recover(); nil != err {
			log.Printf("%v\r\n", err)
		}
		conn.Close()
	}()

	for {
		bytes := make([]byte, *maxPDUSize)
		length, err := conn.Read(bytes)
		if nil != err {
			log.Printf("%v\r\n", err)
			break
		}

		func(buf []byte) {
			client.Send(func() { client.handleRecv(buf) })
		}(bytes[:length])
	}
}

func (client *UdpClient) handleRecv(bytes []byte) {
	pdu, err := DecodePDU(bytes)
	if 0 == pdu.GetRequestID() && nil != err {
		log.Printf("%v\r\n", err)
		return
	}

	req, ok := client.pendings[pdu.GetRequestID()]
	if !ok {
		log.Printf("not found request whit requestId = %d.\r\n", pdu.GetRequestID())
		return
	}

	req.ctx.Reply(pdu, err)
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

	client.requestId++
	pdu.SetRequestID(client.requestId)

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

	client.pendings[pdu.GetRequestID()] = &Request{pdu: pdu, ctx: ctx}

	if *logPdu {
		log.Printf("snmp - " + pdu.String())
	}

	_, err = client.conn.Write(bytes)
	if nil != err {
		err = errors.New("send pdu failed - " + err.Error())
		goto failed
	}

	return
failed:
	ctx.Reply(nil, err)
	return
}

func (client *UdpClient) FreePDU(pdus ...PDU) {

}
