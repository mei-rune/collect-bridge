package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"snmp"
	"strings"
	"time"
	"web"
)

var (
	snmpTimeout = flag.Int("snmp.timeout", 5*60, "maximun duration (second) of send/recv pdu timeout")
)

type SnmpBridge struct {
	snmp.ClientManager
}

func registerSNMP(svr *web.Server) {
	bridge := new(SnmpBridge)
	bridge.Init()
	bridge.Start()
	svr.Get("/snmp/get/(.*)/(.*)", func(ctx *web.Context, host, oid string) { bridge.Get(ctx, host, oid) })
	svr.Put("/snmp/set/(.*)/(.*)", func(ctx *web.Context, host, oid string) { bridge.Set(ctx, host, oid) })
	svr.Get("/snmp/next/(.*)/(.*)", func(ctx *web.Context, host, oid string) { bridge.Next(ctx, host, oid) })
	svr.Get("/snmp/bulk/(.*)/(.*)", func(ctx *web.Context, host, oids string) { bridge.Bulk(ctx, host, oids) })
	svr.Get("/snmp/table/(.*)/(.*)", func(ctx *web.Context, host, oid string) { bridge.Table(ctx, host, oid) })
}

func getTimeout(params map[string]string) time.Duration {
	v, ok := params["timeout"]
	if !ok {
		return time.Duration(*snmpTimeout) * time.Second
	}
	ret, err := snmp.ParseTime(v)
	if nil != err {
		panic(err)
	}
	return ret
}

func getVersion(params map[string]string) snmp.SnmpVersion {
	v, ok := params["version"]
	if !ok {
		return snmp.SNMP_V2C
	}
	switch v {
	case "v1", "V1":
		return snmp.SNMP_V1
	case "v2", "V2", "v2c", "V2C":
		return snmp.SNMP_V2C
	case "v3", "V3":
		return snmp.SNMP_V3
	}
	panic(errors.New(fmt.Sprintf("error version: %s", v)))
	return 0
}

func internalError(ctx *web.Context, msg string, err error) {
	ctx.Abort(500, msg+"\r\n"+err.Error())
}

func (bridge *SnmpBridge) Get(ctx *web.Context, host, oid string) {
	client, err := bridge.GetClient(host)
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	req, err := client.CreatePDU(snmp.SNMP_PDU_GET, getVersion(ctx.Params))
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	err = req.Init(ctx.Params)
	if nil != err {
		internalError(ctx, "init pdu failed.", err)
		return
	}
	err = req.GetVariableBindings().Append(oid, "")
	if nil != err {
		internalError(ctx, "append vb failed.", err)
		return
	}
	resp, ok := client.SendAndRecv(req, getTimeout(ctx.Params))
	if nil == ok && 0 < resp.GetVariableBindings().Len() {
		ctx.WriteString(resp.GetVariableBindings().Get(0).Value.String())
	} else {
		internalError(ctx, "snmp failed.", ok)
	}
	client.FreePDU(req, resp)
}

func (bridge *SnmpBridge) Set(ctx *web.Context, host, oid string) {
	client, err := bridge.GetClient(host)
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	req, err := client.CreatePDU(snmp.SNMP_PDU_SET, getVersion(ctx.Params))
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	err = req.Init(ctx.Params)
	if nil != err {
		internalError(ctx, "init pdu failed.", err)
		return
	}
	txt, err := ioutil.ReadAll(ctx.Request.Body)
	if nil != err {
		internalError(ctx, "read from body failed.", err)
		return
	}

	err = req.GetVariableBindings().Append(oid, string(txt))
	if nil != err {
		internalError(ctx, "append vb failed.", err)
		return
	}

	resp, ok := client.SendAndRecv(req, getTimeout(ctx.Params))
	if nil == ok {
		ctx.WriteString("ok")
	} else {
		internalError(ctx, "snmp failed.", ok)
	}
	client.FreePDU(req, resp)
}

func (bridge *SnmpBridge) Next(ctx *web.Context, host, oid string) {
	client, err := bridge.GetClient(host)
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	req, err := client.CreatePDU(snmp.SNMP_PDU_GETNEXT, getVersion(ctx.Params))
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	err = req.Init(ctx.Params)
	if nil != err {
		internalError(ctx, "init pdu failed.", err)
		return
	}
	err = req.GetVariableBindings().Append(oid, "")
	if nil != err {
		internalError(ctx, "append vb failed.", err)
		return
	}

	resp, ok := client.SendAndRecv(req, getTimeout(ctx.Params))
	if nil == ok {
		ctx.WriteString(resp.GetVariableBindings().Get(0).Value.String())
	} else {
		internalError(ctx, "snmp failed.", ok)
	}
	client.FreePDU(req, resp)
}

func (bridge *SnmpBridge) Bulk(ctx *web.Context, host, oids string) {
	client, err := bridge.GetClient(host)
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	req, err := client.CreatePDU(snmp.SNMP_PDU_GETBULK, getVersion(ctx.Params))
	if nil != err {
		internalError(ctx, "create pdu failed.", err)
		return
	}
	err = req.Init(ctx.Params)
	if nil != err {
		internalError(ctx, "init pdu failed.", err)
		return
	}
	vbs := req.GetVariableBindings()
	for _, oid := range strings.Split(oids, "|") {
		err = vbs.Append(oid, "")
		if nil != err {
			internalError(ctx, "append vb failed.", err)
			return
		}

	}
	resp, ok := client.SendAndRecv(req, getTimeout(ctx.Params))
	switch {
	case nil != ok:
		internalError(ctx, "snmp failed.", ok)
	case resp.GetVariableBindings().Len() > 0:
		var out bytes.Buffer
		out.WriteString("[")
		for _, vb := range resp.GetVariableBindings().All() {
			out.WriteString("\"")
			out.WriteString(vb.Oid.String())
			out.WriteString("\":\"")
			out.WriteString(vb.Value.String())
			out.WriteString("\",")
		}
		out.Truncate(out.Len() - 1)
		out.WriteString("]")
		out.WriteTo(ctx)
	default:
		ctx.WriteString("[]")
	}
	client.FreePDU(req, resp)
}

func (client *SnmpBridge) Table(ctx *web.Context, host, oids string) {
	panic(fmt.Sprintf("not implemented"))
}
