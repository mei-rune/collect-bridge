package snmp

const (
	SNMP_Verr = 0
	SNMP_V1   = 1
	SNMP_V2C  = 2
	SNMP_V3   = 3
)

const (
	SNMP_AUTH_NOAUTH   = 0
	SNMP_AUTH_HMAC_MD5 = 1
	SNMP_AUTH_HMAC_SHA = 2
)

const (
	SNMP_PRIV_NOPRIV = 0
	SNMP_PRIV_DES    = 1
	SNMP_PRIV_AES    = 2
)

const (
	SNMP_PDU_GET      = 0
	SNMP_PDU_GETNEXT  = 1
	SNMP_PDU_RESPONSE = 2
	SNMP_PDU_SET      = 3
	SNMP_PDU_TRAP     = 4 /* v1 */
	SNMP_PDU_GETBULK  = 5 /* v2 */
	SNMP_PDU_INFORM   = 6 /* v2 */
	SNMP_PDU_TRAP2    = 7 /* v2 */
	SNMP_PDU_REPORT   = 8 /* v2 */

	SNMP_TRAP_COLDSTART              = 0
	SNMP_TRAP_WARMSTART              = 1
	SNMP_TRAP_LINKDOWN               = 2
	SNMP_TRAP_LINKUP                 = 3
	SNMP_TRAP_AUTHENTICATION_FAILURE = 4
	SNMP_TRAP_EGP_NEIGHBOR_LOSS      = 5
	SNMP_TRAP_ENTERPRISE             = 6

	SNMP_SECMODEL_ANY     = 0
	SNMP_SECMODEL_SNMPv1  = 1
	SNMP_SECMODEL_SNMPv2c = 2
	SNMP_SECMODEL_USM     = 3
	SNMP_SECMODEL_UNKNOWN = 4
)

///////////////////////// VariableBindings ///////////////////////////////////
type VariableBinding struct {
	Oid   SnmpOid
	Value SnmpValue
}

type VariableBindings struct {
	values []VariableBinding
}

func (vbs *VariableBindings) All() []VariableBinding {
	return vbs.values
}

func (vbs *VariableBindings) Len() int {
	return len(vbs.values)
}

func (vbs *VariableBindings) Get(idx int) VariableBinding {
	return vbs.values[idx]
}

func (vbs *VariableBindings) Put(idx int, oid, value string) error {
	o, ok := ParseOidFromString(oid)
	if nil != ok {
		return ok
	}

	v, ok := NewSnmpValue(value)
	if nil != ok {
		return ok
	}

	vbs.values[idx].Oid = o
	vbs.values[idx].Value = v
	return nil
}

func (vbs *VariableBindings) Append(oid, value string) error {
	o, ok := ParseOidFromString(oid)
	if nil != ok {
		return ok
	}

	v, ok := NewSnmpValue(value)
	if nil != ok {
		return ok
	}

	if nil == vbs.values {
		vbs.values = make([]VariableBinding, 0, 20)
	}
	vbs.values = append(vbs.values, VariableBinding{Oid: o, Value: v})
	return nil
}

func (vbs *VariableBindings) AppendWith(oid SnmpOid, value SnmpValue) error {
	if nil == vbs.values {
		vbs.values = make([]VariableBinding, 0, 20)
	}
	vbs.values = append(vbs.values, VariableBinding{Oid: oid, Value: value})
	return nil
}

type PDU interface {
	Init(params map[string]string) error
	GetRequestID() int
	GetVersion() int
	GetType() int
	GetTarget() string
	GetVariableBindings() *VariableBindings
}

type SecurityModel interface {
	Init(params map[string]string) error
}

type snmpEngine struct {
	engine_id    []byte
	engine_boots int
	engine_time  int
	max_msg_size uint
}

type Client interface {
	CreatePDU(op, version int) (PDU, error)
	SendAndRecv(req PDU) (PDU, error)
	FreePDU(pdus ...PDU)
}
