package snmp

// #include "bsnmp/config.h"
// #include "bsnmp/asn1.h"
// #include "bsnmp/snmp.h"
// #include "bsnmp/gobindings.h"
//  
// #cgo CFLAGS: -O0 -g3
// #cgo LDFLAGS: -lws2_32
import "C"

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"unsafe"
)

const (
	MAX_COMMUNITY_LEN     = 128
	SNMP_ENGINE_ID_LEN    = 32
	SNMP_CONTEXT_NAME_LEN = 32
	SNMP_AUTH_KEY_LEN     = 40
	SNMP_PRIV_KEY_LEN     = 32
	SNMP_ADM_STR32_LEN    = 32
)

type V2CPDU struct {
	version          SnmpVersion
	op               SnmpType
	requestId        int
	target           string
	community        string
	variableBindings VariableBindings
	client           Client
}

func (pdu *V2CPDU) Init(params map[string]string) error {
	community, ok := params["community"]
	if ok {
		pdu.community = community
	}
	return nil
}

func (pdu *V2CPDU) GetRequestID() int {
	return pdu.requestId
}

func (pdu *V2CPDU) SetRequestID(id int) {
	pdu.requestId = id
}

func (pdu *V2CPDU) GetVersion() SnmpVersion {
	return pdu.version
}

func (pdu *V2CPDU) GetType() SnmpType {
	return pdu.op
}

func (pdu *V2CPDU) GetTarget() string {
	return pdu.target
}

func (pdu *V2CPDU) GetVariableBindings() *VariableBindings {
	return &pdu.variableBindings
}

func (pdu *V2CPDU) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(pdu.op.String())
	buffer.WriteString(" variableBindings")
	buffer.WriteString(pdu.variableBindings.String())
	buffer.WriteString(" from ")
	buffer.WriteString(pdu.target)
	buffer.WriteString(" with community = '")
	buffer.WriteString(pdu.community)
	buffer.WriteString("' and requestId='")
	buffer.WriteString(strconv.Itoa(pdu.GetRequestID()))
	buffer.WriteString("' and version='")
	buffer.WriteString(pdu.version.String())
	buffer.WriteString("'")
	return buffer.String()
}

func (pdu *V2CPDU) encodePDU() ([]byte, error) {
	var internal C.snmp_pdu_t
	C.snmp_pdu_init(&internal)
	defer C.snmp_pdu_free(&internal)

	err := strcpy(&internal.community[0], MAX_COMMUNITY_LEN, pdu.community)
	if nil != err {
		return nil, err
	}
	internal.engine.max_msg_size = 10000
	internal.request_id = C.int32_t(pdu.requestId)
	internal.pdu_type = C.u_int(pdu.op)
	internal.version = uint32(pdu.version)

	err = encodeBindings(&internal, pdu.GetVariableBindings())

	if nil != err {
		return nil, err
	}

	return encodeNativePdu(&internal)
}

func (pdu *V2CPDU) decodePDU(native *C.snmp_pdu_t) error {
	native.community[MAX_COMMUNITY_LEN-1] = 0
	pdu.community = C.GoString(&native.community[0])

	pdu.requestId = int(native.request_id)
	pdu.op = SnmpType(native.pdu_type)
	pdu.version = SnmpVersion(native.version)

	decodeBindings(native, pdu.GetVariableBindings())
	return nil
}

type V3PDU struct {
	op               SnmpType
	requestId        int
	target           string
	securityModel    SecurityModel
	variableBindings VariableBindings
	client           Client
	maxMsgSize       uint
	contextName      string
	contextEngine    []byte
}

func (pdu *V3PDU) Init(params map[string]string) (err error) {
	pdu.maxMsgSize = 10000

	if v, ok := params["max_msg_size"]; ok {
		if num, e := strconv.ParseUint(v, 10, 0); nil == e {
			pdu.maxMsgSize = uint(num)
		}
	}

	if s, ok := params["context_name"]; ok {
		pdu.contextName = s
		if s, ok = params["context_engine"]; ok {
			pdu.contextEngine, err = hex.DecodeString(s)
			if nil != err {
				return fmt.Errorf("'context_engine' decode failed, %s", err.Error())
			}
		}
	}
	switch params["secmodel"] {
	case "usm", "Usm", "USM":
		pdu.securityModel = new(USM)
		err = pdu.securityModel.Init(params)
	case "hashusm", "HashUsm", "HASHUSM":
		pdu.securityModel = new(HashUSM)
		err = pdu.securityModel.Init(params)
	default:
		err = errors.New(fmt.Sprintf("unsupported security module: %s", params["secmodel"]))
	}
	return
}

func (pdu *V3PDU) GetRequestID() int {
	return pdu.requestId
}

func (pdu *V3PDU) SetRequestID(id int) {
	pdu.requestId = id
}

func (pdu *V3PDU) GetVersion() SnmpVersion {
	return SNMP_V3
}

func (pdu *V3PDU) GetType() SnmpType {
	return pdu.op
}

func (pdu *V3PDU) GetTarget() string {
	return pdu.target
}

func (pdu *V3PDU) GetVariableBindings() *VariableBindings {
	return &pdu.variableBindings
}

func (pdu *V3PDU) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(pdu.op.String())
	buffer.WriteString(" variableBindings")
	buffer.WriteString(pdu.variableBindings.String())
	buffer.WriteString(" from ")
	buffer.WriteString(pdu.target)
	buffer.WriteString(" with ")
	buffer.WriteString(pdu.securityModel.String())
	buffer.WriteString(" and contextName='")
	buffer.WriteString(pdu.contextName)
	buffer.WriteString("' and contextEngine='")
	buffer.WriteString(hex.EncodeToString(pdu.contextEngine))
	buffer.WriteString(" and ")
	buffer.WriteString(pdu.securityModel.String())
	buffer.WriteString(" and requestId='")
	buffer.WriteString(strconv.Itoa(pdu.GetRequestID()))
	buffer.WriteString("' and version='v3'")
	return buffer.String()
}

func (pdu *V3PDU) encodePDU() ([]byte, error) {
	var internal C.snmp_pdu_t
	C.snmp_pdu_init(&internal)
	defer C.snmp_pdu_free(&internal)

	internal.request_id = C.int32_t(pdu.requestId)
	internal.pdu_type = C.u_int(pdu.op)
	internal.version = uint32(SNMP_V3)

	internal.identifier = C.int32_t(pdu.requestId)
	if 0 == pdu.maxMsgSize {
		pdu.maxMsgSize = 10000
	}
	internal.engine.max_msg_size = C.int32_t(pdu.maxMsgSize)
	internal.flags = 0

	internal.security_model = SNMP_SECMODEL_USM

	err := encodeBindings(&internal, pdu.GetVariableBindings())

	if nil != err {
		return nil, err
	}

	return encodeNativePdu(&internal)
}

func (pdu *V3PDU) decodePDU(native *C.snmp_pdu_t) error {
	//return decodeBindings(native, pdu.GetVariableBindings())
	return errors.New("v3pdu.decodePdu() not implemented")
}

///////////////////////// Encode/Decode /////////////////////////////

func strcpy(dst *C.char, capacity int, src string) error {
	if capacity < len(src) {
		return errors.New("string too long.")
	}
	s := C.CString(src)
	C.strcpy(dst, s)
	C.free(unsafe.Pointer(s))
	return nil
}

const (
	ASN_MAXOIDLEN     = 128
	SNMP_MAX_BINDINGS = 100
)

func oidWrite(dst *C.asn_oid_t, value SnmpValue) error {
	uintArray := value.GetUint32s()
	if ASN_MAXOIDLEN <= len(uintArray) {
		return fmt.Errorf("oid is too long, maximum size is %d, oid is %s", ASN_MAXOIDLEN, value.String())
	}

	for i, subOid := range uintArray {
		dst.subs[i] = C.asn_subid_t(subOid)
	}
	dst.len = C.u_int(len(uintArray))
	return nil
}

func oidRead(src *C.asn_oid_t) *SnmpOid {
	subs := make([]uint32, src.len)
	for i := 0; i < int(src.len); i++ {
		subs[i] = uint32(src.subs[i])
	}
	return NewOid(subs)
}

func encodeNativePdu(pdu *C.snmp_pdu_t) ([]byte, error) {
	bytes := make([]byte, int(pdu.engine.max_msg_size))
	var buffer C.asn_buf_t
	C.set_asn_u_ptr(&buffer.asn_u, (*C.char)(unsafe.Pointer(&bytes[0])))
	buffer.asn_len = C.size_t(len(bytes))

	ret_code := C.snmp_pdu_encode(pdu, &buffer)
	if 0 != ret_code {
		err := errors.New(C.GoString(C.snmp_get_error(ret_code)))
		return nil, err
	}
	length := C.get_buffer_length(&buffer, (*C.u_char)(unsafe.Pointer(&bytes[0])))
	return bytes[0:length], nil
}

func encodeBindings(internal *C.snmp_pdu_t, vbs *VariableBindings) error {

	if SNMP_MAX_BINDINGS < vbs.Len() {
		return fmt.Errorf("bindings too long, SNMP_MAX_BINDINGS is %d, variableBindings is %d",
			SNMP_MAX_BINDINGS, vbs.Len())
	}

	for i, vb := range vbs.All() {
		err := oidWrite(&internal.bindings[i].oid, &vb.Oid)
		if nil != err {
			return err
		}

		if nil == vb.Value {
			internal.bindings[i].syntax = uint32(SNMP_SYNTAX_NULL)
			continue
		}

		internal.bindings[i].syntax = uint32(vb.Value.GetSyntax())
		switch vb.Value.GetSyntax() {
		case SNMP_SYNTAX_NULL:
		case SNMP_SYNTAX_INTEGER:
			C.snmp_value_put_int32(&internal.bindings[i].v, C.int32_t(vb.Value.GetInt32()))
		case SNMP_SYNTAX_OCTETSTRING:
			bytes := vb.Value.GetBytes()
			C.snmp_value_put_octets(&internal.bindings[i].v, unsafe.Pointer(&bytes[0]), C.u_int(len(bytes)))
		case SNMP_SYNTAX_OID:
			err = oidWrite(C.snmp_value_get_oid(&internal.bindings[i].v), vb.Value)
			if nil != err {
				return err
			}
		case SNMP_SYNTAX_IPADDRESS:
			bytes := vb.Value.GetBytes()
			if 4 != len(bytes) {
				return fmt.Errorf("ip address is error, it's length is %d, excepted length is 4, value is %s",
					len(bytes), vb.Value.String())
			}
			C.snmp_value_put_ipaddress(&internal.bindings[i].v, C.u_char(bytes[0]),
				C.u_char(bytes[1]), C.u_char(bytes[2]), C.u_char(bytes[3]))
		case SNMP_SYNTAX_COUNTER:
			C.snmp_value_put_uint32(&internal.bindings[i].v, C.uint32_t(vb.Value.GetUint32()))
		case SNMP_SYNTAX_GAUGE:
			C.snmp_value_put_uint32(&internal.bindings[i].v, C.uint32_t(vb.Value.GetUint32()))
		case SNMP_SYNTAX_TIMETICKS:
			C.snmp_value_put_uint32(&internal.bindings[i].v, C.uint32_t(vb.Value.GetUint32()))
		case SNMP_SYNTAX_COUNTER64:
			C.snmp_value_put_uint64(&internal.bindings[i].v, C.uint64_t(vb.Value.GetUint64()))
		default:
			return fmt.Errorf("unsupported type - %v", vb.Value)
		}
	}
	internal.nbindings = C.u_int(vbs.Len())
	return nil
}

func decodeBindings(internal *C.snmp_pdu_t, vbs *VariableBindings) {

	for i := 0; i < int(internal.nbindings); i++ {
		oid := *oidRead(&internal.bindings[i].oid)

		switch SnmpSyntax(internal.bindings[i].syntax) {
		case SNMP_SYNTAX_NULL:
			vbs.AppendWith(oid, NewSnmpNil())
		case SNMP_SYNTAX_INTEGER:
			vbs.AppendWith(oid, NewSnmpInt32(int32(C.snmp_value_get_int32(&internal.bindings[i].v))))
		case SNMP_SYNTAX_OCTETSTRING:
			l := int(C.snmp_value_get_octets_len(&internal.bindings[i].v))
			bytes := make([]byte, l, l+10)
			C.snmp_value_get_octets(&internal.bindings[i].v, unsafe.Pointer(&bytes[0]))
			vbs.AppendWith(oid, NewSnmpOctetString(bytes))
		case SNMP_SYNTAX_OID:
			v := oidRead(C.snmp_value_get_oid(&internal.bindings[i].v))
			vbs.AppendWith(oid, v)
		case SNMP_SYNTAX_IPADDRESS:
			bytes := make([]byte, 4)
			tmp := C.snmp_value_get_ipaddress(&internal.bindings[i].v)
			C.memcpy(unsafe.Pointer(&bytes[0]), unsafe.Pointer(tmp), 4)
			vbs.AppendWith(oid, NewSnmpAddress(bytes))
		case SNMP_SYNTAX_COUNTER:
			vbs.AppendWith(oid, NewSnmpCounter32(uint32(C.snmp_value_get_uint32(&internal.bindings[i].v))))
		case SNMP_SYNTAX_GAUGE:
			vbs.AppendWith(oid, NewSnmpUint32(uint32(C.snmp_value_get_uint32(&internal.bindings[i].v))))
		case SNMP_SYNTAX_TIMETICKS:
			vbs.AppendWith(oid, NewSnmpTimeticks(uint32(C.snmp_value_get_uint32(&internal.bindings[i].v))))
		case SNMP_SYNTAX_COUNTER64:
			vbs.AppendWith(oid, NewSnmpCounter64(uint64(C.snmp_value_get_uint64(&internal.bindings[i].v))))
		default:
			vbs.AppendWith(oid, NewSnmpError(uint(internal.bindings[i].syntax)))
		}
	}
}

func DecodePDU(bytes []byte) (PDU, error) {
	var buffer C.asn_buf_t
	var pdu C.snmp_pdu_t
	var recv_len C.int32_t

	C.set_asn_u_ptr(&buffer.asn_u, (*C.char)(unsafe.Pointer(&bytes[0])))
	buffer.asn_len = C.size_t(len(bytes))

	C.snmp_pdu_init(&pdu)
	ret_code := C.snmp_pdu_decode(&buffer, &pdu, &recv_len)
	if 0 != ret_code {
		err := errors.New(C.GoString(C.snmp_get_error(ret_code)))
		return nil, err
	}
	defer C.snmp_pdu_free(&pdu)

	if uint32(SNMP_V3) == pdu.version {
		var v3 V3PDU
		return &v3, v3.decodePDU(&pdu)
	}
	var v2 V2CPDU
	return &v2, v2.decodePDU(&pdu)
}

func EncodePDU(pdu PDU) ([]byte, error) {
	if pdu.GetVersion() != SNMP_V3 {
		return pdu.(*V2CPDU).encodePDU()
	}
	return pdu.(*V3PDU).encodePDU()
}
