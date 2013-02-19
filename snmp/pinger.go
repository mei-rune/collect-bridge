// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package snmp

// #include "bsnmp/config.h"
// #include "bsnmp/asn1.h"
// #include "bsnmp/snmp.h"
// #include "bsnmp/gobindings.h"
import "C"
import (
	"commons"
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"
)

type PingResult struct {
	Addr    net.Addr
	Version SnmpVersion
	Error   error
}

type Pinger struct {
	network string
	id      int
	conn    net.PacketConn
	wait    sync.WaitGroup
	ch      chan *PingResult
}

func NewPinger(network, laddr string, capacity int) (*Pinger, error) {
	c, err := net.ListenPacket(network, laddr)
	if err != nil {
		return nil, fmt.Errorf("ListenPacket(%q, %q) failed: %v", network, laddr, err)
	}

	pinger := &Pinger{network: network, id: 1, conn: c,
		ch: make(chan *PingResult, capacity)}

	go pinger.serve()
	pinger.wait.Add(1)
	return pinger, nil
}

func (self *Pinger) Close() {
	self.conn.Close()
	self.conn = nil
	self.wait.Wait()
	close(self.ch)
}

func (self *Pinger) GetChannel() <-chan *PingResult {
	return self.ch
}

func (self *Pinger) Send(raddr string, version SnmpVersion) error {
	ra, err := net.ResolveUDPAddr(self.network, raddr)
	if err != nil {
		return fmt.Errorf("ResolveIPAddr(%q, %q) failed: %v", self.network, raddr, err)
	}

	self.id++

	var pdu PDU = nil
	switch version {
	case SNMP_V1, SNMP_V2C:
		pdu = &V2CPDU{op: SNMP_PDU_GET, version: version, requestId: self.id, target: raddr, community: "public"}
		err = pdu.GetVariableBindings().Append("1.3.6.1.2.1.1.2.0", "")
		if err != nil {
			return fmt.Errorf("AppendVariableBinding failed: %v", err)
		}
	case SNMP_V3:
		pdu = &V3PDU{op: SNMP_PDU_GET, requestId: self.id, identifier: self.id, target: raddr,
			securityModel: &USM{auth_proto: SNMP_AUTH_NOAUTH, priv_proto: SNMP_PRIV_NOPRIV}}
	default:
		return fmt.Errorf("Unsupported version - %v", version)
	}

	bytes, e := EncodePDU(pdu, false)
	if e != nil {
		return fmt.Errorf("EncodePDU failed: %v", e)
	}
	l, err := self.conn.WriteTo(bytes, ra)
	if err != nil {
		return fmt.Errorf("WriteTo failed: %v", err)
	}
	if l == 0 {
		return fmt.Errorf("WriteTo failed: wlen == 0")
	}
	return nil
}

func (self *Pinger) Recv(timeout time.Duration) (net.Addr, SnmpVersion, error) {
	select {
	case res := <-self.ch:
		return res.Addr, res.Version, res.Error
	case <-time.After(timeout):
		return nil, SNMP_Verr, commons.TimeoutErr
	}
	return nil, SNMP_Verr, commons.TimeoutErr
}

func (self *Pinger) serve() {
	defer self.wait.Done()

	reply := make([]byte, 2048)
	var buffer C.asn_buf_t
	var pdu C.snmp_pdu_t

	for nil != self.conn {
		l, ra, err := self.conn.ReadFrom(reply)
		if err != nil {
			self.ch <- &PingResult{Error: fmt.Errorf("ReadFrom failed: %v", err)}
			break
		}

		C.set_asn_u_ptr(&buffer.asn_u, (*C.char)(unsafe.Pointer(&reply[0])))
		buffer.asn_len = C.size_t(l)

		err = DecodePDUHeader(&buffer, &pdu)
		if nil != err {
			self.ch <- &PingResult{Error: fmt.Errorf("Parse Data failed: %s %v", ra.String(), err)}
			continue
		}
		ver := SNMP_Verr
		switch pdu.version {
		case uint32(SNMP_V3):
			ver = SNMP_V3
		case uint32(SNMP_V2C):
			ver = SNMP_V2C
		case uint32(SNMP_V1):
			ver = SNMP_V1
		}
		self.ch <- &PingResult{Addr: ra, Version: ver}
		C.snmp_pdu_free(&pdu)
	}
}
