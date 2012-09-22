package snmp

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type SnmpSyntax uint

const (
	/* v1 additions */
	SNMP_SYNTAX_NULL        SnmpSyntax = 0
	SNMP_SYNTAX_INTEGER     SnmpSyntax = 1 /* == INTEGER32 */
	SNMP_SYNTAX_OCTETSTRING SnmpSyntax = 2
	SNMP_SYNTAX_OID         SnmpSyntax = 3
	SNMP_SYNTAX_IPADDRESS   SnmpSyntax = 4
	SNMP_SYNTAX_COUNTER     SnmpSyntax = 5
	SNMP_SYNTAX_GAUGE       SnmpSyntax = 6 /* == UNSIGNED32 */
	SNMP_SYNTAX_TIMETICKS   SnmpSyntax = 7

	/* v2 additions */
	SNMP_SYNTAX_COUNTER64      SnmpSyntax = 8
	SNMP_SYNTAX_NOSUCHOBJECT   SnmpSyntax = 9  /* exception */
	SNMP_SYNTAX_NOSUCHINSTANCE SnmpSyntax = 10 /* exception */
	SNMP_SYNTAX_ENDOFMIBVIEW   SnmpSyntax = 1  /* exception */
)

type SnmpValue interface {
	IsNil() bool
	GetSyntax() SnmpSyntax
	String() string
	GetUint32s() []uint32
	GetBytes() []byte
	GetInt32() int32
	GetUint32() uint32
	GetUint64() uint64
	GetString() string

	IsError() bool
	Error() string
}

const (
	syntexErrorMessage = "snmp value format error, excepted format is '[type]value'," +
		" type is 'null, int32, gauge, counter32, counter64, octet, oid, ip, timeticks', value is a string. - %s"

	notError = "this is not a error. please call IsError() first."
)

func NewSnmpValue(s string) (SnmpValue, error) {
	if "" == s {
		return nil, fmt.Errorf("input parameter is empty.")
	}
	if s[0] != '[' {
		return nil, fmt.Errorf(syntexErrorMessage, s)
	}
	ss := strings.SplitN(s[1:], "]", 2)
	if 2 != len(ss) {
		return nil, fmt.Errorf(syntexErrorMessage, s)
	}

	switch ss[0] {
	case "null", "Null", "NULL", "nil", "Nil":
		return NewSnmpNil(), nil
	case "int", "int32", "Int", "Int32", "INT", "INT32":
		// error pattern: return newSnmpInt32FromString(ss[1])
		// see http://www.golang.org/doc/go_faq.html#nil_error
		return newSnmpInt32FromString(ss[1])
	case "uint", "uint32", "Uint", "Uint32", "UINT", "UINT32", "gauge", "Gauge", "GAUGE":
		return newSnmpUint32FromString(ss[1])
	case "counter", "Counter", "COUNTER", "counter32", "Counter32", "COUNTER32":
		return newSnmpCounter32FromString(ss[1])
	case "counter64", "Counter64", "COUNTER64":
		return newSnmpCounter64FromString(ss[1])
	case "str", "string", "String", "octetstring", "OctetString", "octets", "Octets", "OCTETS":
		return newSnmpOctetStringFromString(ss[1])
	case "oid", "Oid", "OID":
		return newSnmpOidFromString(ss[1])
	case "ip", "IP", "ipaddress", "IPAddress", "IpAddress":
		return newSnmpAddressFromString(ss[1])
	case "timeticks", "Timeticks", "TIMETICKS":
		return newSnmpTimeticksFromString(ss[1])
	}

	return nil, fmt.Errorf("unsupported snmp type -", ss[0])
}

type SnmpNil struct{}

var (
	snmpNil = new(SnmpNil)
)

func NewSnmpNil() SnmpValue {
	return snmpNil
}

func (oid *SnmpNil) String() string {
	return "[null]"
}

func (s *SnmpNil) IsNil() bool {
	return true
}

func (s *SnmpNil) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_NULL
}

func (s *SnmpNil) GetUint32s() []uint32 {
	return nil
}

func (s *SnmpNil) GetBytes() []byte {
	return nil
}

func (s *SnmpNil) GetInt32() int32 {
	return 0
}

func (s *SnmpNil) GetUint32() uint32 {
	return 0
}

func (s *SnmpNil) GetUint64() uint64 {
	return 0
}

func (s *SnmpNil) GetString() string {
	return ""
}

func (s *SnmpNil) IsError() bool {
	return false
}

func (s *SnmpNil) Error() string {
	return notError
}

type SnmpOid []uint32

func (oid *SnmpOid) String() string {
	return "[oid]" + oid.GetString()
}

func (s *SnmpOid) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_OID
}

func (s *SnmpOid) IsNil() bool {
	return false
}

func (s *SnmpOid) GetUint32s() []uint32 {
	return []uint32(*s)
}

func (s *SnmpOid) GetBytes() []byte {
	return nil
}

func (s *SnmpOid) GetInt32() int32 {
	return 0
}

func (s *SnmpOid) GetUint32() uint32 {
	return 0
}

func (s *SnmpOid) GetUint64() uint64 {
	return 0
}

func (oid *SnmpOid) GetString() string {
	if 0 == len(*oid) {
		return ""
	}

	var result bytes.Buffer
	for _, v := range *oid {
		result.WriteString(strconv.FormatUint(uint64(v), 10))
		result.WriteString(".")
	}

	result.Truncate(result.Len() - 1)
	return result.String()
}

func (s *SnmpOid) IsError() bool {
	return false
}

func (s *SnmpOid) Error() string {
	return notError
}

func NewOid(subs []uint32) *SnmpOid {
	ret := SnmpOid(subs)
	return &ret
}

func ParseOidFromString(s string) (SnmpOid, error) {
	result := make([]uint32, 0, 20)
	ss := strings.Split(s, ".")
	if 2 > len(ss) {
		ss = strings.Split(s, "_")
	}
	for i, v := range ss {
		if 0 == len(v) {
			if 0 != i {
				return nil, fmt.Errorf("oid style error, value is %s", s)
			}
		} else {
			num, ok := strconv.ParseUint(v, 10, 0)
			if nil != ok {
				return nil, fmt.Errorf("oid style error, value is %s, exception is $s", s, ok.Error())
			}

			result = append(result, uint32(num))
		}
	}
	return SnmpOid(result), nil
}

func newSnmpOidFromString(s string) (SnmpValue, error) {
	oid, error := ParseOidFromString(s)
	if nil == error {
		return &oid, error
	}
	return nil, error
}

type SnmpInt32 int32

func (v *SnmpInt32) String() string {
	return "[int32]" + v.GetString()
}

func (v *SnmpInt32) IsNil() bool {
	return false
}

func (v *SnmpInt32) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_INTEGER
}

func (v *SnmpInt32) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpInt32) GetBytes() []byte {
	return nil
}

func (v *SnmpInt32) GetInt32() int32 {
	return int32(*v)
}

func (v *SnmpInt32) GetUint32() uint32 {
	return uint32(*v)
}

func (v *SnmpInt32) GetUint64() uint64 {
	return uint64(*v)
}

func (v *SnmpInt32) GetString() string {
	return strconv.FormatInt(int64(*v), 10)
}

func (s *SnmpInt32) IsError() bool {
	return false
}

func (s *SnmpInt32) Error() string {
	return notError
}

func NewSnmpInt32(v int32) *SnmpInt32 {
	ret := SnmpInt32(v)
	return &ret
}

func newSnmpInt32FromString(s string) (SnmpValue, error) {
	i, ok := strconv.ParseInt(s, 10, 0)
	if nil != ok {
		return nil, fmt.Errorf("int32 style error, value is %s, exception is %s", s, ok.Error())
	}
	var ret SnmpInt32 = SnmpInt32(i)
	return &ret, nil
}

type SnmpUint32 uint32

func (v *SnmpUint32) String() string {
	return "[gauge]" + v.GetString()
}

func (s *SnmpUint32) IsNil() bool {
	return false
}

func (s *SnmpUint32) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_GAUGE
}

func (v *SnmpUint32) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpUint32) GetBytes() []byte {
	return nil
}

func (v *SnmpUint32) GetInt32() int32 {
	return int32(*v)
}

func (v *SnmpUint32) GetUint32() uint32 {
	return uint32(*v)
}

func (v *SnmpUint32) GetUint64() uint64 {
	return uint64(*v)
}

func (v *SnmpUint32) GetString() string {
	return strconv.FormatUint(uint64(*v), 10)
}

func (s *SnmpUint32) IsError() bool {
	return false
}

func (s *SnmpUint32) Error() string {
	return notError
}

func NewSnmpUint32(v uint32) *SnmpUint32 {
	ret := SnmpUint32(v)
	return &ret
}

func newSnmpUint32FromString(s string) (SnmpValue, error) {
	i, ok := strconv.ParseUint(s, 10, 0)
	if nil != ok {
		return nil, fmt.Errorf("gauge style error, value is %s, exception is %s", s, ok.Error())
	}
	var ret SnmpUint32 = SnmpUint32(i)
	return &ret, nil
}

type SnmpCounter32 uint32

func (v *SnmpCounter32) String() string {
	return "[counter32]" + v.GetString()
}

func (s *SnmpCounter32) IsNil() bool {
	return false
}

func (s *SnmpCounter32) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_COUNTER
}

func (v *SnmpCounter32) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpCounter32) GetBytes() []byte {
	return nil
}

func (v *SnmpCounter32) GetInt32() int32 {
	return int32(*v)
}

func (v *SnmpCounter32) GetUint32() uint32 {
	return uint32(*v)
}

func (v *SnmpCounter32) GetUint64() uint64 {
	return uint64(*v)
}

func (v *SnmpCounter32) GetString() string {
	return strconv.FormatUint(uint64(*v), 10)
}

func (s *SnmpCounter32) IsError() bool {
	return false
}

func (s *SnmpCounter32) Error() string {
	return notError
}

func NewSnmpCounter32(v uint32) *SnmpCounter32 {
	ret := SnmpCounter32(v)
	return &ret
}

func newSnmpCounter32FromString(s string) (SnmpValue, error) {
	i, ok := strconv.ParseUint(s, 10, 0)
	if nil != ok {
		return nil, fmt.Errorf("counter32 style error, value is %s, exception is %s", s, ok.Error())
	}
	var ret SnmpCounter32 = SnmpCounter32(i)
	return &ret, nil
}

type SnmpCounter64 uint64

func (v *SnmpCounter64) String() string {
	return "[counter64]" + v.GetString()
}

func (s *SnmpCounter64) IsNil() bool {
	return false
}

func (s *SnmpCounter64) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_COUNTER64
}

func (v *SnmpCounter64) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpCounter64) GetBytes() []byte {
	return nil
}

func (v *SnmpCounter64) GetInt32() int32 {
	return int32(*v)
}

func (v *SnmpCounter64) GetUint32() uint32 {
	return uint32(*v)
}

func (v *SnmpCounter64) GetUint64() uint64 {
	return uint64(*v)
}

func (v *SnmpCounter64) GetString() string {
	return strconv.FormatUint(uint64(*v), 10)
}

func (s *SnmpCounter64) IsError() bool {
	return false
}

func (s *SnmpCounter64) Error() string {
	return notError
}

func NewSnmpCounter64(v uint64) *SnmpCounter64 {
	ret := SnmpCounter64(v)
	return &ret
}

func newSnmpCounter64FromString(s string) (SnmpValue, error) {
	i, ok := strconv.ParseUint(s, 10, 64)
	if nil != ok {
		return nil, fmt.Errorf("counter64 style error, value is %s, exception is %s", s, ok.Error())
	}
	var ret SnmpCounter64 = SnmpCounter64(i)
	return &ret, nil
}

type SnmpTimeticks uint32

func (v *SnmpTimeticks) String() string {
	return "[timeticks]" + v.GetString()
}

func (s *SnmpTimeticks) IsNil() bool {
	return false
}

func (s *SnmpTimeticks) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_TIMETICKS
}

func (v *SnmpTimeticks) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpTimeticks) GetBytes() []byte {
	return nil
}

func (v *SnmpTimeticks) GetInt32() int32 {
	return int32(*v)
}

func (v *SnmpTimeticks) GetUint32() uint32 {
	return uint32(*v)
}

func (v *SnmpTimeticks) GetUint64() uint64 {
	return uint64(*v)
}

func (v *SnmpTimeticks) GetString() string {
	return strconv.FormatUint(uint64(*v), 10)
}

func (s *SnmpTimeticks) IsError() bool {
	return false
}

func (s *SnmpTimeticks) Error() string {
	return notError
}

func NewSnmpTimeticks(v uint32) *SnmpTimeticks {
	ret := SnmpTimeticks(v)
	return &ret
}

func newSnmpTimeticksFromString(s string) (SnmpValue, error) {
	i, ok := strconv.ParseUint(s, 10, 64)
	if nil != ok {
		return nil, fmt.Errorf("snmpTimeticks style error, value is %s, exception is %s", s, ok.Error())
	}
	var ret SnmpTimeticks = SnmpTimeticks(i)
	return &ret, nil
}

type SnmpOctetString []byte

func (v *SnmpOctetString) String() string {
	return "[octets]" + v.GetString()
}

func (s *SnmpOctetString) IsNil() bool {
	return false
}

func (s *SnmpOctetString) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_OCTETSTRING
}

func (v *SnmpOctetString) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpOctetString) GetBytes() []byte {
	return []byte(*v)
}

func (v *SnmpOctetString) GetInt32() int32 {
	r, e := strconv.ParseInt(string(*v), 10, 32)
	if nil != e {
		return 0
	}
	return int32(r)
}

func (v *SnmpOctetString) GetUint32() uint32 {
	return uint32(v.GetUint64())
}

func (v *SnmpOctetString) GetUint64() uint64 {
	r, e := strconv.ParseUint(string(*v), 10, 64)
	if nil != e {
		return 0
	}
	return r
}

func (v *SnmpOctetString) GetString() string {
	return string(*v)
}

func (s *SnmpOctetString) IsError() bool {
	return false
}

func (s *SnmpOctetString) Error() string {
	return notError
}

func NewSnmpOctetString(v []byte) *SnmpOctetString {
	ret := SnmpOctetString(v)
	return &ret
}

func newSnmpOctetStringFromString(s string) (SnmpValue, error) {
	var ret SnmpOctetString = SnmpOctetString([]byte(s))
	return &ret, nil
}

type SnmpAddress net.IP

func (v *SnmpAddress) String() string {
	return "[ip]" + v.GetString()
}

func (s *SnmpAddress) IsNil() bool {
	return false
}

func (s *SnmpAddress) GetSyntax() SnmpSyntax {
	return SNMP_SYNTAX_IPADDRESS
}

func (v *SnmpAddress) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpAddress) GetBytes() []byte {
	ip := net.IP(*v)
	bytes := ip.To4()
	if nil == bytes {
		return []byte(*v)
	}
	return []byte(bytes)
}

func (v *SnmpAddress) GetInt32() int32 {
	return 0
}

func (v *SnmpAddress) GetUint32() uint32 {
	return 0
}

func (v *SnmpAddress) GetUint64() uint64 {
	return 0
}

func (v *SnmpAddress) GetString() string {
	return net.IP(*v).String()
}

func (s *SnmpAddress) IsError() bool {
	return false
}

func (s *SnmpAddress) Error() string {
	return notError
}

func NewSnmpAddress(v []byte) *SnmpAddress {
	ret := SnmpAddress(net.IP(v))
	return &ret
}

func newSnmpAddressFromString(s string) (SnmpValue, error) {
	addr := net.ParseIP(s)
	if nil == addr {
		return nil, fmt.Errorf("SnmpAddress style error, value is %s", s)
	}
	sa := SnmpAddress(addr)
	return &sa, nil
}

type SnmpError struct {
	value   SnmpSyntax
	message string
}

func (v *SnmpError) String() string {
	return "[error:" + strconv.Itoa(int(v.value)) + "]" + v.message
}

func (s *SnmpError) IsNil() bool {
	return false
}

func (s *SnmpError) GetSyntax() SnmpSyntax {
	return s.value
}

func (v *SnmpError) GetUint32s() []uint32 {
	return nil
}

func (v *SnmpError) GetBytes() []byte {
	return nil
}

func (v *SnmpError) GetInt32() int32 {
	return 0
}

func (v *SnmpError) GetUint32() uint32 {
	return 0
}

func (v *SnmpError) GetUint64() uint64 {
	return 0
}

func (v *SnmpError) GetString() string {
	return ""
}

func (s *SnmpError) IsError() bool {
	return true
}

func (s *SnmpError) Error() string {
	return s.message
}

func errorToMessage(value uint) string {
	switch SnmpSyntax(value) {
	case SNMP_SYNTAX_NOSUCHOBJECT:
		return "nosuchobject"
	case SNMP_SYNTAX_NOSUCHINSTANCE:
		return "nosuchinstance"
	case SNMP_SYNTAX_ENDOFMIBVIEW:
		return "endofmibview"
	}
	return "unknown_snmp_syntax_" + strconv.FormatUint(uint64(value), 10)
}

func NewSnmpError(value uint) *SnmpError {
	return &SnmpError{value: SnmpSyntax(value), message: errorToMessage(value)}
}

func NewSnmpErrorWithMessage(value uint, err string) *SnmpError {
	return &SnmpError{value: SnmpSyntax(value), message: err}
}
