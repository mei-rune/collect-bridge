package discovery

type SnmpV3Params struct {
	AuthProto string `json:"auth_proto"`
	PrivProto string `json:"priv_proto"`

	AuthPassphrase string `json:"auth_pass"`
	PrivPassphrase string `json:"priv_pass"`

	ContextName   string `json:"ctx_name"`
	ContextEngine string `json:"ctx_engine"`
}

type DiscoveryParams struct {
	Network      []string `json:"network"`
	Communities  []string `json:"communities"`
	SnmpV3Params []string `json:"snmpv3_params"`
	Depth        int      `json:"discovery_depth"`
}
