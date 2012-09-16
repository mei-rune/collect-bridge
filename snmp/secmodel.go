package snmp

import (
	"errors"
	"strings"
)

type HashUSM struct {
	auth_proto int
	priv_proto int
	auth_key   []byte
	priv_key   []byte
	name       string
}

func getAuth(params map[string]string) (int, string, error) {
	auth, ok := params["auth_pass"]

	if !ok {
		return SNMP_AUTH_NOAUTH, "", nil
	}
	ss := strings.SplitN(auth, "-", 2)
	if 2 != len(ss) {
		return SNMP_AUTH_NOAUTH, "", errors.New("auth passphrase hasn`t auth protocol. " +
			"please input auth key with \"protocol-passphrase\", auth protocol is \"md5\" or \"sha\"")
	}

	switch ss[0] {
	case "md5", "MD5":
		return SNMP_AUTH_HMAC_MD5, ss[1], nil
	case "sha", "SHA":
		return SNMP_AUTH_HMAC_SHA, ss[1], nil
	}
	return SNMP_AUTH_NOAUTH, "", errors.New("unsupported auth protocol. " +
		"auth protocol must is \"md5\" or \"sha\"")
}

func getPriv(params map[string]string) (int, string, error) {
	priv, ok := params["priv_pass"]

	if !ok {
		return SNMP_PRIV_NOPRIV, "", nil
	}

	ss := strings.SplitN(priv, "-", 2)
	if 2 != len(ss) {
		return SNMP_PRIV_NOPRIV, "", errors.New("priv passphrase hasn`t priv protocol. " +
			"please input priv key with \"protocol-passphrase\", priv protocol is \"des\" or \"aes\"")
	}

	switch ss[0] {
	case "des", "DES":
		return SNMP_PRIV_DES, ss[1], nil
	case "aes", "AES":
		return SNMP_PRIV_AES, ss[1], nil
	}
	return SNMP_PRIV_NOPRIV, "", errors.New("unsupported priv protocol. " +
		"priv protocol must is \"des\" or \"aes\"")
}

func (usm *HashUSM) Init(params map[string]string) error {
	name, ok := params["secname"]
	if !ok {
		return errors.New("secname is required.")
	}
	usm.name = name

	proto, value, err := getAuth(params)
	if nil != err {
		return err
	}

	usm.auth_proto = proto
	usm.auth_key = []byte(value)

	proto, value, err = getPriv(params)
	if nil != err {
		return err
	}

	usm.priv_proto = proto
	usm.priv_key = []byte(value)
	return nil
}

type USM struct {
	auth_proto int
	priv_proto int
	auth_key   string
	priv_key   string
	name       string
}

func (usm *USM) Init(params map[string]string) error {
	name, ok := params["secname"]
	if !ok {
		return errors.New("secname is required.")
	}
	usm.name = name

	proto, value, err := getAuth(params)
	if nil != err {
		return err
	}

	usm.auth_proto = proto
	usm.auth_key = value

	proto, value, err = getPriv(params)
	if nil != err {
		return err
	}

	usm.priv_proto = proto
	usm.priv_key = value
	return nil
}
