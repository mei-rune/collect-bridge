package snmp

import (
	"encoding/hex"
	"testing"
)

// GET SNMPv1 '123987' request_id=234 error_status=0 error_index=0
//  [0]: 1.2.3.4.5.1.7.8.9.10.11.12.13=NULL
//  [1]: 1.2.3.4.5.2.7.8.9.10.11.12.13=INTEGER 12
//  [2]: 1.2.3.4.5.3.7.8.9.10.11.12.13=OCTET STRING 10: 31 32 33 34 35 36 37 38 39 30
//  [3]: 1.2.3.4.5.4.7.8.9.10.11.12.13=OID 2.3.4.5.6.7.8.9.10.11.12.13
//  [4]: 1.2.3.4.5.5.7.8.9.10.11.12.13=IPADDRESS 1.2.3.4
//  [5]: 1.2.3.4.5.6.7.8.9.10.11.12.13=COUNTER 2235683
//  [6]: 1.2.3.4.5.7.7.8.9.10.11.12.13=GAUGE 1235683
//  [7]: 1.2.3.4.5.8.7.8.9.10.11.12.13=TIMETICKS 1235683
//  [8]: 1.2.3.4.5.9.7.8.9.10.11.12.13=COUNTER64 12352121212122683
const snmpv1_txt = "3081e70201000406313233393837a081d9020200ea0201000201003081cc3010060c2a030405010708090a0b0c0d05003011060c2a030405020708090a0b0c0d02010c301a060c2a030405030708090a0b0c0d040a31323334353637383930301b060c2a030405040708090a0b0c0d060b530405060708090a0b0c0d3014060c2a030405050708090a0b0c0d4004010203043013060c2a030405060708090a0b0c0d4103221d233013060c2a030405070708090a0b0c0d420312dae33013060c2a030405080708090a0b0c0d430312dae33017060c2a030405090708090a0b0c0d46072be2305512363b"

// GET SNMPv2c '123987' request_id=234 error_status=0 error_index=0
//  [0]: 1.2.3.4.5.1.7.8.9.10.11.12.13=NULL
//  [1]: 1.2.3.4.5.2.7.8.9.10.11.12.13=INTEGER 12
//  [2]: 1.2.3.4.5.3.7.8.9.10.11.12.13=OCTET STRING 10: 31 32 33 34 35 36 37 38 39 30
//  [3]: 1.2.3.4.5.4.7.8.9.10.11.12.13=OID 2.3.4.5.6.7.8.9.10.11.12.13
//  [4]: 1.2.3.4.5.5.7.8.9.10.11.12.13=IPADDRESS 1.2.3.4
//  [5]: 1.2.3.4.5.6.7.8.9.10.11.12.13=COUNTER 2235683
//  [6]: 1.2.3.4.5.7.7.8.9.10.11.12.13=GAUGE 1235683
//  [7]: 1.2.3.4.5.8.7.8.9.10.11.12.13=TIMETICKS 1235683
//  [8]: 1.2.3.4.5.9.7.8.9.10.11.12.13=COUNTER64 12352121212122683
const snmpv2c_txt = "3081e70201010406313233393837a081d9020200ea0201000201003081cc3010060c2a030405010708090a0b0c0d05003011060c2a030405020708090a0b0c0d02010c301a060c2a030405030708090a0b0c0d040a31323334353637383930301b060c2a030405040708090a0b0c0d060b530405060708090a0b0c0d3014060c2a030405050708090a0b0c0d4004010203043013060c2a030405060708090a0b0c0d4103221d233013060c2a030405070708090a0b0c0d420312dae33013060c2a030405080708090a0b0c0d430312dae33017060c2a030405090708090a0b0c0d46072be2305512363b"

// GET SNMPv3 '' request_id=234 error_status=0 error_index=0
//  [0]: 1.2.3.4.5.1.7.8.9.10.11.12.13=NULL
//  [1]: 1.2.3.4.5.2.7.8.9.10.11.12.13=INTEGER 12
//  [2]: 1.2.3.4.5.3.7.8.9.10.11.12.13=OCTET STRING 10: 31 32 33 34 35 36 37 38 39 30
//  [3]: 1.2.3.4.5.4.7.8.9.10.11.12.13=OID 2.3.4.5.6.7.8.9.10.11.12.13
//  [4]: 1.2.3.4.5.5.7.8.9.10.11.12.13=IPADDRESS 1.2.3.4
//  [5]: 1.2.3.4.5.6.7.8.9.10.11.12.13=COUNTER 2235683
//  [6]: 1.2.3.4.5.7.7.8.9.10.11.12.13=GAUGE 1235683
//  [7]: 1.2.3.4.5.8.7.8.9.10.11.12.13=TIMETICKS 1235683
//  [8]: 1.2.3.4.5.9.7.8.9.10.11.12.13=COUNTER64 12352121212122683
const snmpv3_noauth_txt = "30820150020103300d020100020227170401040201030438303604203031323334353637383930313233343536373839303132333435363738393031020103020204d204076d65696a696e670400040030820100041174657374636f6e74657874656e67696e65040f74657374636f6e746578746e616d65a081d9020200ea0201000201003081cc3010060c2a030405010708090a0b0c0d05003011060c2a030405020708090a0b0c0d02010c301a060c2a030405030708090a0b0c0d040a31323334353637383930301b060c2a030405040708090a0b0c0d060b530405060708090a0b0c0d3014060c2a030405050708090a0b0c0d4004010203043013060c2a030405060708090a0b0c0d4103221d233013060c2a030405070708090a0b0c0d420312dae33013060c2a030405080708090a0b0c0d430312dae33017060c2a030405090708090a0b0c0d46072be2305512363b"

// GET SNMPv3 '' request_id=234 error_status=0 error_index=0
//  [0]: 1.2.3.4.5.1.7.8.9.10.11.12.13=NULL
//  [1]: 1.2.3.4.5.2.7.8.9.10.11.12.13=INTEGER 12
//  [2]: 1.2.3.4.5.3.7.8.9.10.11.12.13=OCTET STRING 10: 31 32 33 34 35 36 37 38 39 30
//  [3]: 1.2.3.4.5.4.7.8.9.10.11.12.13=OID 2.3.4.5.6.7.8.9.10.11.12.13
//  [4]: 1.2.3.4.5.5.7.8.9.10.11.12.13=IPADDRESS 1.2.3.4
//  [5]: 1.2.3.4.5.6.7.8.9.10.11.12.13=COUNTER 2235683
//  [6]: 1.2.3.4.5.7.7.8.9.10.11.12.13=GAUGE 1235683
//  [7]: 1.2.3.4.5.8.7.8.9.10.11.12.13=TIMETICKS 1235683
//  [8]: 1.2.3.4.5.9.7.8.9.10.11.12.13=COUNTER64 12352121212122683
const snmpv3_md5_txt = "3082015c020103300d020100020227170401050201030444304204203031323334353637383930313233343536373839303132333435363738393031020103020204d204076d65696a696e67040c3ecad6303ab094cf9fc49cc8040030820100041174657374636f6e74657874656e67696e65040f74657374636f6e746578746e616d65a081d9020200ea0201000201003081cc3010060c2a030405010708090a0b0c0d05003011060c2a030405020708090a0b0c0d02010c301a060c2a030405030708090a0b0c0d040a31323334353637383930301b060c2a030405040708090a0b0c0d060b530405060708090a0b0c0d3014060c2a030405050708090a0b0c0d4004010203043013060c2a030405060708090a0b0c0d4103221d233013060c2a030405070708090a0b0c0d420312dae33013060c2a030405080708090a0b0c0d430312dae33017060c2a030405090708090a0b0c0d46072be2305512363b"

// GET SNMPv3 '' request_id=234 error_status=0 error_index=0
//  [0]: 1.2.3.4.5.1.7.8.9.10.11.12.13=NULL
//  [1]: 1.2.3.4.5.2.7.8.9.10.11.12.13=INTEGER 12
//  [2]: 1.2.3.4.5.3.7.8.9.10.11.12.13=OCTET STRING 10: 31 32 33 34 35 36 37 38 39 30
//  [3]: 1.2.3.4.5.4.7.8.9.10.11.12.13=OID 2.3.4.5.6.7.8.9.10.11.12.13
//  [4]: 1.2.3.4.5.5.7.8.9.10.11.12.13=IPADDRESS 1.2.3.4
//  [5]: 1.2.3.4.5.6.7.8.9.10.11.12.13=COUNTER 2235683
//  [6]: 1.2.3.4.5.7.7.8.9.10.11.12.13=GAUGE 1235683
//  [7]: 1.2.3.4.5.8.7.8.9.10.11.12.13=TIMETICKS 1235683
//  [8]: 1.2.3.4.5.9.7.8.9.10.11.12.13=COUNTER64 12352121212122683
const snmpv3_md5_des_txt = "3082016c020103300d02010002022717040107020103044c304a04203031323334353637383930313233343536373839303132333435363738393031020103020204d204076d65696a696e67040cc414cae9ec0af879221fe89904080300000029000000048201085d8e848967040c913b715e3ee20c3a175f430e774fc770d5c012e7dcd6207ae331a937ba936b521f858dd89fcec0e86516d22d6993c5b369d2df77309abe6c1e61af12305272737684b0edac7f3e9029a22fd538aa725192217133731f5e50cec6ccaf14b3a90ad688001f4cc88a10cf14aab9168ef6e8d136192af95655ef6e030325ec04a7bd0067deff5a9b9239c51c7b9adcdd9b4d3c3069cc13efe4e8535d3c2982b63f41f0da79fc920b9bf0e01886b5e7f3da222298ce15834dddf494169b71874489c981154582cfdb5f5df9815c25e788dd4a90edc0a96ca8eeae7aaebe4e9109fedec7faf1a983c5893767383d7e16a0bccef02f14a781c382ec6b24637d1fa1a3f401"

// GET SNMPv3 '' request_id=234 error_status=0 error_index=0
//  [0]: 1.2.3.4.5.1.7.8.9.10.11.12.13=NULL
//  [1]: 1.2.3.4.5.2.7.8.9.10.11.12.13=INTEGER 12
//  [2]: 1.2.3.4.5.3.7.8.9.10.11.12.13=OCTET STRING 10: 31 32 33 34 35 36 37 38 39 30
//  [3]: 1.2.3.4.5.4.7.8.9.10.11.12.13=OID 2.3.4.5.6.7.8.9.10.11.12.13
//  [4]: 1.2.3.4.5.5.7.8.9.10.11.12.13=IPADDRESS 1.2.3.4
//  [5]: 1.2.3.4.5.6.7.8.9.10.11.12.13=COUNTER 2235683
//  [6]: 1.2.3.4.5.7.7.8.9.10.11.12.13=GAUGE 1235683
//  [7]: 1.2.3.4.5.8.7.8.9.10.11.12.13=TIMETICKS 1235683
//  [8]: 1.2.3.4.5.9.7.8.9.10.11.12.13=COUNTER64 12352121212122683
const snmpv3_sha_txt = "3082015c020103300d020100020227170401050201030444304204203031323334353637383930313233343536373839303132333435363738393031020103020204d204076d65696a696e67040ce7a696149d5fd4e6fdb17cd9040030820100041174657374636f6e74657874656e67696e65040f74657374636f6e746578746e616d65a081d9020200ea0201000201003081cc3010060c2a030405010708090a0b0c0d05003011060c2a030405020708090a0b0c0d02010c301a060c2a030405030708090a0b0c0d040a31323334353637383930301b060c2a030405040708090a0b0c0d060b530405060708090a0b0c0d3014060c2a030405050708090a0b0c0d4004010203043013060c2a030405060708090a0b0c0d4103221d233013060c2a030405070708090a0b0c0d420312dae33013060c2a030405080708090a0b0c0d430312dae33017060c2a030405090708090a0b0c0d46072be2305512363b"

// GET SNMPv3 '' request_id=234 error_status=0 error_index=0
//  [0]: 1.2.3.4.5.1.7.8.9.10.11.12.13=NULL
//  [1]: 1.2.3.4.5.2.7.8.9.10.11.12.13=INTEGER 12
//  [2]: 1.2.3.4.5.3.7.8.9.10.11.12.13=OCTET STRING 10: 31 32 33 34 35 36 37 38 39 30
//  [3]: 1.2.3.4.5.4.7.8.9.10.11.12.13=OID 2.3.4.5.6.7.8.9.10.11.12.13
//  [4]: 1.2.3.4.5.5.7.8.9.10.11.12.13=IPADDRESS 1.2.3.4
//  [5]: 1.2.3.4.5.6.7.8.9.10.11.12.13=COUNTER 2235683
//  [6]: 1.2.3.4.5.7.7.8.9.10.11.12.13=GAUGE 1235683
//  [7]: 1.2.3.4.5.8.7.8.9.10.11.12.13=TIMETICKS 1235683
//  [8]: 1.2.3.4.5.9.7.8.9.10.11.12.13=COUNTER64 12352121212122683
const snmpv3_sha_aes_txt = "30820168020103300d02010002022717040107020103044c304a04203031323334353637383930313233343536373839303132333435363738393031020103020204d204076d65696a696e67040ca3f43fa5687d10f27616544c040823480000be1800000482010479ab3546d6732de5704f3aa5fd37f650f027932db936963781dda6ab507bd814a5f3ba65fb68ef394f7028f899487492e76855130d50059042a2f7c59a686849b8d510eabbf1d9fa5f9968535c80a60540bbe1985a2f78810549a2fa8bffedcdf827eb8976f7dbc14266394adaba3569dc1974c0003b4602c9c2909c768d871ab6d9d3ea892cab901990cc547367e0853dd99cb3a871bdc22eefa50f573107edcd9eefbce827cd20fc370589ddd14eebc8be629884bd0af384fee99c1b1eaf3c03e12e5c70ed00dae9caf7eabcca8f22ab10b0d7e6374412db478091c62bf46d0b25a4048e4ecd57b890b1122a385b49eb3aa6306abfda33e19e76bdbe0ef8dea06f0c40"

const snmpv2c_NOSUCHINSTANCE = "302002010104067075626c6963a2130201010201050201013008300606022b068100"

var oid1 []uint32 = []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
var oid2 []uint32 = []uint32{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

// void append_bindings(snmp_pdu_t* pdu, asn_subid_t* oid
// 	, u_int oid_len, enum snmp_syntax syntax ) {

// 	pdu->bindings[pdu->nbindings].var.len = oid_len;
// 	memcpy(pdu->bindings[pdu->nbindings].var.subs, oid, oid_len*sizeof(oid[0]));
// 	pdu->bindings[pdu->nbindings].syntax = syntax;
// 	pdu->bindings[pdu->nbindings].var.subs[5] = pdu->nbindings + 1;
// 	pdu->nbindings ++;
// }

func AppendBindings(vbs *VariableBindings, s string) {
	oid := make([]uint32, len(oid1))
	copy(oid, oid1)
	oid[5] = uint32(vbs.Len() + 1)
	value, e := NewSnmpValue(s)
	if nil != e {
		panic(e)
	}
	vbs.AppendWith(SnmpOid(oid), value)
}

func checkOid(target *SnmpOid, i int, t *testing.T) {
	oid := make([]uint32, len(oid1))
	copy(oid, oid1)
	oid[5] = uint32(i + 1)
	if target.String() != NewOid(oid).String() {
		t.Errorf("decode v1 pdu faile - oid[%d] not equal, excepted is %s, value is %s", i, NewOid(oid).GetString(), target.GetString())
	}
}
func uint32ToString(ints []uint32) string {
	oid := SnmpOid(ints)
	return oid.String()
}

func fillPdu(vbs *VariableBindings) {
	AppendBindings(vbs, "[null]")
	AppendBindings(vbs, "[int32]12")
	AppendBindings(vbs, "[octets]1234567890")
	AppendBindings(vbs, uint32ToString(oid2))
	AppendBindings(vbs, "[ip]1.2.3.4")
	AppendBindings(vbs, "[counter32]2235683")
	AppendBindings(vbs, "[gauge]1235683")
	AppendBindings(vbs, "[timeticks]1235683")
	AppendBindings(vbs, "[counter64]12352121212122683")
}

func checkVB(vbs *VariableBindings, i int, excepted string, t *testing.T) {
	oid := vbs.Get(i).Oid
	checkOid(&oid, i, t)
	if vbs.Get(i).Value.String() != excepted {
		t.Errorf("decode v1 pdu faile - value[%d] error, excepted is '%s', value is %s", i, excepted, vbs.Get(i).Value.String())
	}
}

func checkPdu(vbs *VariableBindings, t *testing.T) {
	checkVB(vbs, 0, "[null]", t)
	checkVB(vbs, 1, "[int32]12", t)
	checkVB(vbs, 2, "[octets]1234567890", t)
	checkVB(vbs, 3, uint32ToString(oid2), t)
	checkVB(vbs, 4, "[ip]1.2.3.4", t)
	checkVB(vbs, 5, "[counter32]2235683", t)
	checkVB(vbs, 6, "[gauge]1235683", t)
	checkVB(vbs, 7, "[timeticks]1235683", t)
	checkVB(vbs, 8, "[counter64]12352121212122683", t)
}

func TestEncodePDU(t *testing.T) {
	pdu := &V2CPDU{version: SNMP_V1, requestId: 234}
	pdu.Init(map[string]string{"community": "123987"})
	fillPdu(pdu.GetVariableBindings())
	bytes, e := pdu.encodePDU()
	if nil != e {
		t.Errorf("encode v1 pdu faile - %s", e.Error())
	}

	if snmpv1_txt != hex.EncodeToString(bytes) {
		t.Log(hex.EncodeToString(bytes))
		t.Errorf("encode v1 pdu faile.")
	}

	pdu = &V2CPDU{version: SNMP_V2C, requestId: 234}
	pdu.Init(map[string]string{"community": "123987"})
	fillPdu(pdu.GetVariableBindings())
	bytes, e = pdu.encodePDU()
	if nil != e {
		t.Errorf("encode v2 pdu faile - %s", e.Error())
	}

	if snmpv2c_txt != hex.EncodeToString(bytes) {
		t.Log(hex.EncodeToString(bytes))
		t.Errorf("encode v2 pdu faile.")
	}

}

func TestDecodePDU(t *testing.T) {
	bytes, e := hex.DecodeString(snmpv1_txt)
	if nil != e {
		t.Errorf("decode hex failed - %s", e.Error())
		return
	}
	pdu, e := DecodePDU(bytes)
	if nil != e {
		t.Errorf("decode v1 pdu faile - %s", e.Error())
	} else {
		if SNMP_V1 != pdu.GetVersion() {
			t.Errorf("decode v1 pdu faile - version error, excepted is v1, actual value is %d", pdu.GetVersion())
		} else {
			if "123987" != pdu.(*V2CPDU).community {
				t.Errorf("decode v1 pdu faile - community error, excepted is '123987', actual value is %s", pdu.(*V2CPDU).community)
			}
			if 234 != pdu.(*V2CPDU).requestId {
				t.Errorf("decode v1 pdu faile - requestId error, excepted is '234', actual value is %d", pdu.(*V2CPDU).requestId)
			}
		}
		checkPdu(pdu.GetVariableBindings(), t)
	}

	bytes, e = hex.DecodeString(snmpv2c_txt)
	if nil != e {
		t.Errorf("decode hex failed - %s", e.Error())
		return
	}
	pdu, e = DecodePDU(bytes)
	if nil != e {
		t.Errorf("decode v1 pdu faile - %s", e.Error())
	} else {
		if SNMP_V2C != pdu.GetVersion() {
			t.Errorf("decode v2 pdu faile - version error, excepted is v2C, actual value is %d", pdu.GetVersion())
		} else {
			if "123987" != pdu.(*V2CPDU).community {
				t.Errorf("decode v1 pdu faile - community error, excepted is '123987', actual value is %s", pdu.(*V2CPDU).community)
			}
			if 234 != pdu.(*V2CPDU).requestId {
				t.Errorf("decode v1 pdu faile - requestId error, excepted is '234', actual value is %d", pdu.(*V2CPDU).requestId)
			}
		}
		checkPdu(pdu.GetVariableBindings(), t)
	}

}
