package netutils

import (
	"bytes"
	"commons"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
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

const (
	ICMP4_ECHO_REQUEST = 8
	ICMP4_ECHO_REPLY   = 0
	ICMP6_ECHO_REQUEST = 128
	ICMP6_ECHO_REPLY   = 129
)

type PingResult struct {
	Addr      net.Addr
	Bytes     []byte
	Err       error
	Timestamp int64
}

type Pinger struct {
	family       int
	network      string
	seqnum       int
	id           int
	echo         []byte
	conn         net.PacketConn
	wait         sync.WaitGroup
	ch           chan *PingResult
	cached_bytes []byte
	newRequest   func(id, seqnum, msglen int, filler, cached []byte) []byte
}

func newPinger(family int, network, laddr string, echo []byte, capacity int) (*Pinger, error) {
	c, err := net.ListenPacket(network, laddr)
	if err != nil {
		return nil, fmt.Errorf("ListenPacket(%q, %q) failed: %v", network, laddr, err)
	}

	if nil == echo || 0 == len(echo) {
		echo = []byte("gogogogo")
	}

	newRequest := newICMPv4EchoRequest
	if family == syscall.AF_INET6 {
		newRequest = newICMPv6EchoRequest
	}

	icmp := &Pinger{family: family,
		network:      network,
		seqnum:       61455,
		id:           os.Getpid() & 0xffff,
		echo:         echo,
		conn:         c,
		ch:           make(chan *PingResult, capacity),
		cached_bytes: make([]byte, 1024),
		newRequest:   newRequest}
	//icmp.Send("127.0.0.1", nil)
	go icmp.serve()
	icmp.wait.Add(1)
	return icmp, nil
}

func NewPinger(netwwork, laddr string, echo []byte, capacity int) (*Pinger, error) {
	if netwwork == "ip4:icmp" {
		return newPinger(syscall.AF_INET, netwwork, laddr, echo, capacity)
	}
	if netwwork == "ip6:icmp" {
		return newPinger(syscall.AF_INET6, netwwork, laddr, echo, capacity)
	}
	return nil, errors.New("Unsupported network - " + netwwork)
}

func (self *Pinger) Close() {
	self.conn.Close()
	self.wait.Wait()
	close(self.ch)
}

func (self *Pinger) GetChannel() <-chan *PingResult {
	return self.ch
}

func (self *Pinger) Send(raddr string, echo []byte) error {
	ra, err := net.ResolveIPAddr(self.network, raddr)
	if err != nil {
		return fmt.Errorf("ResolveIPAddr(%q, %q) failed: %v", self.network, raddr, err)
	}
	return self.SendWithIPAddr(ra, echo)
}

func (self *Pinger) SendWithIPAddr(ra *net.IPAddr, echo []byte) error {
	self.seqnum++
	filler := echo
	if nil == filler {
		filler = self.echo
	}

	msglen := len(filler) + 8
	if msglen > 1024 {
		msglen = 2039
	}

	bytes := self.newRequest(self.id, self.seqnum, msglen, filler, self.cached_bytes)
	_, err := self.conn.WriteTo(bytes, ra)
	if err != nil {
		return fmt.Errorf("WriteTo failed: %v", err)
	}
	return nil
}

func (self *Pinger) Recv(timeout time.Duration) (net.Addr, []byte, error) {
	select {
	case res := <-self.ch:
		return res.Addr, res.Bytes, res.Err
	case <-time.After(timeout):
		return nil, nil, commons.TimeoutErr
	}
	return nil, nil, commons.TimeoutErr
}

func (self *Pinger) serve() {
	defer self.wait.Done()

	for nil != self.conn {
		reply := make([]byte, 2048)
		l, ra, err := self.conn.ReadFrom(reply)
		if err != nil {
			self.ch <- &PingResult{Err: fmt.Errorf("ReadFrom failed: %v", err)}
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
		self.ch <- &PingResult{Addr: ra, Bytes: bytes, Timestamp: time.Now().Unix()}
	}
}

func newICMPEchoRequest(family, id, seqnum, msglen int, filler, cached []byte) []byte {
	if family == syscall.AF_INET6 {
		return newICMPv6EchoRequest(id, seqnum, msglen, filler, cached)
	}
	return newICMPv4EchoRequest(id, seqnum, msglen, filler, cached)
}

func newICMPv4EchoRequest(id, seqnum, msglen int, filler, cached []byte) []byte {
	b := newICMPInfoMessage(id, seqnum, msglen, filler, cached)
	b[0] = ICMP4_ECHO_REQUEST

	// calculate Pinger checksum
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

func newICMPv6EchoRequest(id, seqnum, msglen int, filler, cached []byte) []byte {
	b := newICMPInfoMessage(id, seqnum, msglen, filler, cached)
	b[0] = ICMP6_ECHO_REQUEST
	return b
}

func newICMPInfoMessage(id, seqnum, msglen int, filler, cached []byte) []byte {
	b := cached[0:msglen]
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

type Pingers struct {
	*Pinger
}

func NewPingers(echo []byte, capacity int) (*Pingers, error) {
	v4, e := NewPinger("ip4:icmp", "", echo, capacity)
	if nil != e {
		return nil, e
	}
	return &Pingers{v4}, nil
}
