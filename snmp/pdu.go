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
	version          int
	op               int
	requestId        int
	target           string
	community        string
	variableBindings VariableBindings
	client           Client

	internal C.snmp_pdu_t
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

func (pdu *V2CPDU) GetVersion() int {
	return pdu.version
}

func (pdu *V2CPDU) GetType() int {
	return pdu.op
}

func (pdu *V2CPDU) GetTarget() string {
	return pdu.target
}

func (pdu *V2CPDU) GetVariableBindings() *VariableBindings {
	return &pdu.variableBindings
}

func (pdu *V2CPDU) encodePDU() ([]byte, error) {
	C.snmp_pdu_init(&pdu.internal)
	defer C.snmp_pdu_free(&pdu.internal)
	err := initNativePDUv2(pdu)
	if nil != err {
		return nil, err
	}

	err = snmp_add_binding(&pdu.internal, pdu.GetVariableBindings())

	if nil != err {
		return nil, err
	}

	return encodeNativePdu(&pdu.internal)
}

type V3PDU struct {
	op               int
	requestId        int
	target           string
	securityModel    SecurityModel
	variableBindings VariableBindings
	client           Client
	maxMsgSize       uint
	contextName      string
	contextEngine    []byte

	internal C.snmp_pdu_t
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

func (pdu *V3PDU) GetVersion() int {
	return SNMP_V3
}

func (pdu *V3PDU) GetType() int {
	return pdu.op
}

func (pdu *V3PDU) GetTarget() string {
	return pdu.target
}

func (pdu *V3PDU) GetVariableBindings() *VariableBindings {
	return &pdu.variableBindings
}

func (pdu *V3PDU) encodePDU() ([]byte, error) {
	C.snmp_pdu_init(&pdu.internal)
	defer C.snmp_pdu_free(&pdu.internal)

	err := initNativePDUv3(pdu)
	if nil != err {
		return nil, err
	}

	err = snmp_add_binding(&pdu.internal, pdu.GetVariableBindings())

	if nil != err {
		return nil, err
	}

	return encodeNativePdu(&pdu.internal)
}

///////////////////////// Encode/Decode /////////////////////////////
func DecodePDU(bytes []byte) (PDU, error) {
	return nil, errors.New("not implemented")
}

func EncodePDU(pdu PDU) ([]byte, error) {
	if pdu.GetVersion() != SNMP_V3 {
		return pdu.(*V2CPDU).encodePDU()
	}
	return pdu.(*V3PDU).encodePDU()
}

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

func oidCopy(dst *C.asn_oid_t, value SnmpValue) error {
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
func snmp_add_binding(internal *C.snmp_pdu_t, vbs *VariableBindings) error {

	if SNMP_MAX_BINDINGS < vbs.Len() {
		return fmt.Errorf("bindings too long, SNMP_MAX_BINDINGS is %d, variableBindings is %d",
			SNMP_MAX_BINDINGS, vbs.Len())
	}

	for i, vb := range vbs.All() {
		err := oidCopy(&internal.bindings[i].oid, &vb.Oid)
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
			err = oidCopy(C.snmp_value_get_oid(&internal.bindings[i].v), vb.Value)
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
			C.snmp_value_put_uint32(&internal.bindings[i].v, C.uint32_t(vb.Value.GetUint64()))
		}
	}
	internal.nbindings = C.u_int(vbs.Len())
	return nil
}

func initNativePDUv2(pdu *V2CPDU) (err error) {

	err = strcpy(&pdu.internal.community[0], MAX_COMMUNITY_LEN, pdu.community)
	if nil != err {
		return
	}
	pdu.internal.engine.max_msg_size = 10000

	pdu.internal.request_id = C.int32_t(pdu.requestId)
	pdu.internal.pdu_type = C.u_int(pdu.op)
	pdu.internal.version = uint32(pdu.version)
	pdu.internal.error_status = 0
	pdu.internal.error_index = 0
	pdu.internal.nbindings = 0
	return nil
}

func initNativePDUv3(pdu *V3PDU) (err error) {
	C.snmp_pdu_init(&pdu.internal)

	pdu.internal.pdu_type = C.u_int(pdu.op)
	pdu.internal.version = uint32(SNMP_V3)
	pdu.internal.error_status = 0
	pdu.internal.error_index = 0
	pdu.internal.nbindings = 0

	pdu.internal.identifier = C.int32_t(pdu.requestId)
	if 0 == pdu.maxMsgSize {
		pdu.maxMsgSize = 10000
	}
	pdu.internal.engine.max_msg_size = C.int32_t(pdu.maxMsgSize)
	pdu.internal.flags = 0

	pdu.internal.security_model = SNMP_SECMODEL_USM

	// if SNMP_ENGINE_ID_LEN < len(client.engine.engine_id) {
	// 	err = errors.New("engine id too long.")
	// 	return
	// }

	// C.memcpy(unsafe.Pointer(&pdu.internal.engine.engine_id[0]),
	// 	unsafe.Pointer(&client.engine.engine_id[0]),
	// 	C.size_t(len(client.engine.engine_id)))
	// pdu.internal.engine.engine_len = C.uint32_t(len(client.engine.engine_id))
	// pdu.internal.engine.engine_boots = C.int32_t(client.engine.engine_boots)
	// pdu.internal.engine.engine_time = C.int32_t(client.engine.engine_time)
	// pdu.internal.engine.max_msg_size = C.int32_t(client.engine.max_msg_size)

	// pdu.internal.user.auth_proto = uint32(client.user.auth_proto)
	// pdu.internal.user.priv_proto = uint32(client.user.priv_proto)
	// if SNMP_AUTH_KEY_LEN < len(client.user.auth_key) {
	// 	err = errors.New("auth key too long.")
	// 	return
	// }
	// C.memcpy(unsafe.Pointer(&pdu.internal.user.auth_key),
	// 	unsafe.Pointer(&client.user.auth_key[0]),
	// 	C.size_t(len(client.user.auth_key)))
	// pdu.internal.user.auth_len = C.size_t(len(client.user.auth_key))

	// if SNMP_PRIV_KEY_LEN < len(client.user.priv_key) {
	// 	err = errors.New("priv key too long.")
	// 	return
	// }
	// C.memcpy(unsafe.Pointer(&pdu.internal.user.priv_key),
	// 	unsafe.Pointer(&client.user.priv_key[0]),
	// 	C.size_t(len(client.user.priv_key)))
	// pdu.internal.user.priv_len = C.size_t(len(client.user.priv_key))

	// strcpy(&pdu.internal.user.sec_name[0], SNMP_ADM_STR32_LEN, client.user.name)

	// client.init_secparams(&pdu.internal)

	// if SNMP_ENGINE_ID_LEN < len(client.engine.engine_id) {
	// 	err = errors.New("engine id too long.")
	// 	return
	// }
	// C.memcpy(unsafe.Pointer(&pdu.internal.context_engine[0]),
	// 	unsafe.Pointer(&client.engine.engine_id[0]),
	// 	C.size_t(len(client.engine.engine_id)))

	// pdu.internal.context_engine_len = C.uint32_t(len(client.engine.engine_id))

	// err = strcpy(&pdu.internal.context_name[0],
	// 	SNMP_CONTEXT_NAME_LEN, client.ContextName)
	// if nil != err {
	// 	return
	// }

	//     strlcpy(pdu->context_name, client->ContextName,
	//         sizeof(pdu->context_name));
	return
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
