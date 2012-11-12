package snmp

// #include "bsnmp/config.h"
// #include "bsnmp/asn1.h"
// #include "bsnmp/snmp.h"
// #include "bsnmp/gobindings.h"
import "C"
import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
	"unsafe"
)

var maxPDUSize = flag.Int("maxPDUSize", 10240, "set max size of pdu")
var logPdu = flag.Bool("log.pdu", true, "log pdu content?")

type Request struct {
	pdu      PDU
	ctx      InvokedContext
	callback func(PDU, error)
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

func (client *UdpClient) discoverEngine(fn func(PDU, error)) {
	usm := &USM{auth_proto: SNMP_AUTH_NOAUTH, priv_proto: SNMP_PRIV_NOPRIV}
	pdu := &V3PDU{op: SNMP_PDU_GET, target: client.host, securityModel: usm}
	client.sendPdu(pdu, nil, fn)
}

func (client *UdpClient) sendV3PDU(ctx InvokedContext, pdu *V3PDU) {
	if !pdu.securityModel.IsLocalize() {
		if nil == pdu.engine {
			ctx.Reply(nil, errors.New("nil == pdu.engine"))
			return
		}
		pdu.securityModel.Localize(pdu.engine.engine_id)
	}

	client.sendPdu(pdu, ctx, nil)
}

func (client *UdpClient) discoverEngineAndSend(ctx InvokedContext, pdu *V3PDU) {

	if nil != pdu.engine && nil != pdu.engine.engine_id && 0 != len(pdu.engine.engine_id) {
		client.sendV3PDU(ctx, pdu)
		return
	}

	if nil != client.engine.engine_id && 0 != len(client.engine.engine_id) {

		if nil == pdu.engine {
			pdu.engine = &client.engine
		} else {
			pdu.engine.CopyFrom(&client.engine)
		}
		client.sendV3PDU(ctx, pdu)
		return
	}

	client.discoverEngine(func(resp PDU, err error) {

		if nil != err {
			ctx.Reply(nil, errors.New("discover engine failed - "+err.Error()))
			return
		}
		v3, ok := resp.(*V3PDU)
		if !ok {
			ctx.Reply(nil, errors.New("discover engine failed - oooooooooooo! it is not v3pdu"))
			return
		}

		client.engine.CopyFrom(v3.engine)
		if nil == pdu.engine {
			pdu.engine = &client.engine
		} else {
			pdu.engine.engine_id = client.engine.engine_id
		}
		client.sendV3PDU(ctx, pdu)
	})
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

	var buffer C.asn_buf_t
	var pdu C.snmp_pdu_t
	var result PDU
	var req *Request
	var ok bool

	C.set_asn_u_ptr(&buffer.asn_u, (*C.char)(unsafe.Pointer(&bytes[0])))
	buffer.asn_len = C.size_t(len(bytes))

	err := DecodePDUHeader(&buffer, &pdu)
	if nil != err {
		log.Printf("%v\r\n", err)
		return
	}
	defer C.snmp_pdu_free(&pdu)

	if uint32(SNMP_V3) == pdu.version {

		req, ok = client.pendings[int(pdu.identifier)]
		if !ok {
			log.Printf("not found request with requestId = %d.\r\n", int(pdu.identifier))
			return
		}

		v3old, ok := req.pdu.(*V3PDU)
		if !ok {
			err = errors.New("receive pdu is a v3 pdu.")
			goto complete
		}
		usm, ok := v3old.securityModel.(*USM)
		if !ok {
			err = errors.New("receive pdu is not usm.")
			goto complete
		}
		err = FillUser(&pdu, usm.auth_proto, usm.localization_auth_key,
			usm.priv_proto, usm.localization_priv_key)
		if nil != err {
			log.Printf("%s\r\n", err.Error())
			goto complete
		}

		err = DecodePDUBody(&buffer, &pdu)
		if nil != err {
			log.Printf("%s\r\n", err.Error())
			goto complete
		}
		var v3 V3PDU
		err = v3.decodePDU(&pdu)
		result = &v3
	} else {
		err = DecodePDUBody(&buffer, &pdu)
		if nil != err {
			log.Printf("%s\r\n", err.Error())
			return
		}

		req, ok = client.pendings[int(pdu.request_id)]
		if !ok {
			log.Printf("not found request with requestId = %d.\r\n", int(pdu.request_id))
			return
		}

		var v2 V2CPDU
		err = v2.decodePDU(&pdu)
		result = &v2
	}

complete:
	if nil != req.ctx {
		req.ctx.Reply(result, err)
	} else {
		req.callback(result, err)
	}
}

func (client *UdpClient) handleSend(ctx InvokedContext, pdu PDU) {
	var err error = nil
	if nil == client.conn {
		err = client.createConnect()
		if nil != err {
			goto failed
		}
	}

	if SNMP_V3 == pdu.GetVersion() {
		v3, ok := pdu.(*V3PDU)
		if !ok {
			err = errors.New("oooooooooooo! it is not v3pdu.")
			goto failed
		}

		client.discoverEngineAndSend(ctx, v3)
		return
	}

	client.sendPdu(pdu, ctx, nil)
	return
failed:
	ctx.Reply(nil, err)
	return
}

func (client *UdpClient) sendPdu(pdu PDU, ctx InvokedContext, callback func(PDU, error)) {
	if nil == ctx {
		if nil == callback {
			panic("'ctx' and 'callback' is nil")
		}
	} else if nil != callback {
		panic("'ctx' and 'callback' is not nil")
	}

	var bytes []byte = nil
	var err error = nil
	client.requestId++
	pdu.SetRequestID(client.requestId)

	_, ok := client.pendings[pdu.GetRequestID()]
	if ok {
		err = errors.New("identifier is repected.")
		goto failed
	}

	bytes, err = EncodePDU(pdu)
	if nil != err {
		err = errors.New("encode pdu failed - " + err.Error())
		goto failed
	}

	client.pendings[pdu.GetRequestID()] = &Request{pdu: pdu, ctx: ctx, callback: callback}

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
