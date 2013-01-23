// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nmap

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	//"time"
)

func family(a *net.IPAddr) int {
	if a == nil || len(a.IP) <= net.IPv4len {
		return syscall.AF_INET
	}
	if a.IP.To4() != nil {
		return syscall.AF_INET
	}
	return syscall.AF_INET6
}

type pingResult struct {
	addr  net.Addr
	bytes []byte
	err   error
}

type ICMP struct {
	family     int
	network    string
	seqnum     int
	id         int
	echo       []byte
	conn       net.PacketConn
	wait       sync.WaitGroup
	ch         chan *pingResult
	newRequest func(id, seqnum, msglen int, filler []byte) []byte
}

func newICMP(family int, network, laddr string, echo []byte) (*ICMP, error) {
	c, err := net.ListenPacket(network, laddr)
	if err != nil {
		return nil, fmt.Errorf("ListenPacket(%q, %q) failed: %v", network, laddr, err)
	}

	if nil == echo {
		echo = []byte("gogogogo")
	}

	newRequest := newICMPv4EchoRequest
	if family == syscall.AF_INET6 {
		newRequest = newICMPv6EchoRequest
	}

	icmp := &ICMP{family: family,
		network:    network,
		seqnum:     61455,
		id:         os.Getpid() & 0xffff,
		echo:       echo,
		conn:       c,
		ch:         make(chan *pingResult, 100),
		newRequest: newRequest}
	icmp.Send("127.0.0.1", nil)
	go icmp.serve()
	icmp.wait.Add(1)
	return icmp, nil
}

func NewICMP(netwwork, laddr string, echo []byte) (*ICMP, error) {
	if netwwork == "ip4:icmp" {
		return newICMP(syscall.AF_INET, netwwork, laddr, echo)
	}
	if netwwork == "ip6:icmp" {
		return newICMP(syscall.AF_INET6, netwwork, laddr, echo)
	}
	return nil, fmt.Errorf("Unsupported network - %s", netwwork)
}

func (self *ICMP) Close() {
	self.conn.Close()
	self.wait.Wait()
	defer close(self.ch)
}

func (self *ICMP) Send(raddr string, echo []byte) error {
	self.seqnum++
	filler := echo
	if nil == filler {
		filler = self.echo
	}

	msglen := len(filler) + 8
	if msglen > 1024 {
		msglen = 2039
	}

	bytes := self.newRequest(self.id, self.seqnum, msglen, filler)

	ra, err := net.ResolveIPAddr(self.network, raddr)
	if err != nil {
		return fmt.Errorf("ResolveIPAddr(%q, %q) failed: %v", self.network, raddr, err)
	}

	_, err = self.conn.WriteTo(bytes, ra)
	if err != nil {
		return fmt.Errorf("WriteTo failed: %v", err)
	}
	return nil
}

var timeout_error = errors.New("recv icmp packet time out")

func (self *ICMP) Recv(timeout time.Duration) (net.Addr, []byte, error) {
	select {
	case res := <-self.ch:
		return res.addr, res.bytes, res.err
	case <-time.After(timeout):
		return nil, nil, timeout_error
	}
	return nil, nil, timeout_error
}

func (self *ICMP) serve() {
	defer self.wait.Done()

	for nil != self.conn {
		reply := make([]byte, 2048)
		l, ra, err := self.conn.ReadFrom(reply)
		if err != nil {
			self.ch <- &pingResult{err: fmt.Errorf("ReadFrom failed: %v", err)}
			break
		}

		switch self.family {
		case syscall.AF_INET:
			if reply[0] != ICMP4_ECHO_REPLY {
				continue
			}
		case syscall.AF_INET6:
			if reply[0] != ICMP6_ECHO_REPLY {
				continue
			}
		}
		_, _, _, _, bytes := parseICMPEchoReply(reply[:l])
		self.ch <- &pingResult{addr: ra, bytes: bytes}
	}
}

// func icmpEchoTransponder(t *testing.T, network, raddr string, waitForReady chan bool) {
// 	c, err := net.Dial(network, raddr)
// 	if err != nil {
// 		waitForReady <- true
// 		t.Errorf("Dial(%q, %q) failed: %v", network, raddr, err)
// 		return
// 	}
// 	c.SetDeadline(time.Now().Add(100 * time.Millisecond))
// 	defer c.Close()
// 	waitForReady <- true

// 	echo := make([]byte, 256)
// 	var nr int
// 	for {
// 		nr, err = c.Read(echo)
// 		if err != nil {
// 			t.Errorf("Read failed: %v", err)
// 			return
// 		}
// 		switch family(nil) {
// 		case syscall.AF_INET:
// 			if echo[0] != ICMP4_ECHO_REQUEST {
// 				continue
// 			}
// 		case syscall.AF_INET6:
// 			if echo[0] != ICMP6_ECHO_REQUEST {
// 				continue
// 			}
// 		}
// 		break
// 	}

// 	switch family(c.RemoteAddr()) {
// 	case syscall.AF_INET:
// 		echo[0] = ICMP4_ECHO_REPLY
// 	case syscall.AF_INET6:
// 		echo[0] = ICMP6_ECHO_REPLY
// 	}

// 	_, err = c.Write(echo[:nr])
// 	if err != nil {
// 		t.Errorf("Write failed: %v", err)
// 		return
// 	}
// }

const (
	ICMP4_ECHO_REQUEST = 8
	ICMP4_ECHO_REPLY   = 0
	ICMP6_ECHO_REQUEST = 128
	ICMP6_ECHO_REPLY   = 129
)

func newICMPEchoRequest(family, id, seqnum, msglen int, filler []byte) []byte {
	if family == syscall.AF_INET6 {
		return newICMPv6EchoRequest(id, seqnum, msglen, filler)
	}
	return newICMPv4EchoRequest(id, seqnum, msglen, filler)
}

func newICMPv4EchoRequest(id, seqnum, msglen int, filler []byte) []byte {
	b := newICMPInfoMessage(id, seqnum, msglen, filler)
	b[0] = ICMP4_ECHO_REQUEST

	// calculate ICMP checksum
	cklen := len(b)
	s := uint32(0)
	for i := 0; i < cklen-1; i += 2 {
		s += uint32(b[i+1])<<8 | uint32(b[i])
	}
	if cklen&1 == 1 {
		s += uint32(b[cklen-1])
	}
	s = (s >> 16) + (s & 0xffff)
	s = s + (s >> 16)
	// place checksum back in header; using ^= avoids the
	// assumption the checksum bytes are zero
	b[2] ^= uint8(^s & 0xff)
	b[3] ^= uint8(^s >> 8)

	return b
}

func newICMPv6EchoRequest(id, seqnum, msglen int, filler []byte) []byte {
	b := newICMPInfoMessage(id, seqnum, msglen, filler)
	b[0] = ICMP6_ECHO_REQUEST
	return b
}

func newICMPInfoMessage(id, seqnum, msglen int, filler []byte) []byte {
	b := make([]byte, msglen)
	copy(b[8:], bytes.Repeat(filler, (msglen-8)/len(filler)+1))
	b[0] = 0                    // type
	b[1] = 0                    // code
	b[2] = 0                    // checksum
	b[3] = 0                    // checksum
	b[4] = uint8(id >> 8)       // identifier
	b[5] = uint8(id & 0xff)     // identifier
	b[6] = uint8(seqnum >> 8)   // sequence number
	b[7] = uint8(seqnum & 0xff) // sequence number
	return b
}

func parseICMPEchoReply(b []byte) (t, code, id, seqnum int, body []byte) {
	t = int(b[0])
	code = int(b[1])
	id = int(b[4])<<8 | int(b[5])
	seqnum = int(b[6])<<8 | int(b[7])
	return t, code, id, seqnum, b[8:]
}
