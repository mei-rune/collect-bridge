package discovery

import ()

type Interface struct {
	Index        int    `json:"index"`            // positive integer that starts at one, zero is never used
	MTU          int    `json:"mtu"`              // maximum transmission unit
	Description  string `json:"description"`      // e.g., "en0", "lo0", "eth0.100"
	Address      string `json:"address"`          // IP Address
	HardwareAddr string `json:"hardware_address"` // IEEE MAC-48, EUI-48 and EUI-64 form
}

type Device struct {
	ManagedIP   string   `json:"managed_ip"`
	Communities []string `json:"communities"`

	Interfaces []Interface            `json:"interfaces"`
	Attributes map[string]interface{} `json:"attributes"`
}
