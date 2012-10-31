package snmp

// #include "bsnmp/config.h"
// #include "bsnmp/asn1.h"
// #include "bsnmp/snmp.h"
import "C"

import (
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type securityModelWithCopy interface {
	SecurityModel
	Write(*C.snmp_user_t) error
	Read(*C.snmp_user_t) error
}

func getAuth(params map[string]string) (crypto.Hash, AuthType, string, error) {
	auth, ok := params["auth_pass"]

	if !ok {
		return 0, SNMP_AUTH_NOAUTH, "", nil
	}
	ss := strings.SplitN(auth, "-", 2)
	if 2 != len(ss) {
		return 0, SNMP_AUTH_NOAUTH, "", errors.New("auth passphrase hasn`t auth protocol. " +
			"please input auth key with \"protocol-passphrase\", auth protocol is \"md5\" or \"sha\"")
	}

	switch ss[0] {
	case "md5", "MD5":
		return crypto.MD5, SNMP_AUTH_HMAC_MD5, ss[1], nil
	case "sha", "SHA":
		return crypto.SHA1, SNMP_AUTH_HMAC_SHA, ss[1], nil
	}
	return 0, SNMP_AUTH_NOAUTH, "", errors.New("unsupported auth protocol. " +
		"auth protocol must is \"md5\" or \"sha\"")
}

func getPriv(params map[string]string) (PrivType, string, error) {
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

func NewSecurityModel(params map[string]string) (sm securityModelWithCopy, err error) {
	switch params["secmodel"] {
	case "usm", "Usm", "USM":
		securityModel := new(USM)
		securityModel.InitString(params)
		sm = securityModel
	case "hashusm", "HashUsm", "HASHUSM":
		securityModel := new(USM)
		securityModel.InitHash(params)
		sm = securityModel
	default:
		err = errors.New(fmt.Sprintf("unsupported security module: %s", params["secmodel"]))
	}
	return
}

type USM struct {
	hash            crypto.Hash
	auth_proto      AuthType
	priv_proto      PrivType
	auth_passphrase string
	priv_passphrase string
	auth_key        []byte
	priv_key        []byte

	localization_auth_key []byte
	localization_priv_key []byte

	name string
}

func (usm *USM) InitHash(params map[string]string) error {
	name, ok := params["secname"]
	if !ok {
		return errors.New("secname is required.")
	}
	usm.name = name

	hash, auth_proto, value, err := getAuth(params)
	if nil != err {
		return err
	}
	usm.hash = hash
	usm.auth_proto = auth_proto
	usm.auth_key = []byte(value)

	priv_proto, value, err := getPriv(params)
	if nil != err {
		return err
	}

	usm.priv_proto = priv_proto
	usm.priv_key = []byte(value)
	return nil
}

func (usm *USM) InitString(params map[string]string) error {
	name, ok := params["secname"]
	if !ok {
		return errors.New("secname is required.")
	}
	usm.name = name

	hash, auth_proto, value, err := getAuth(params)
	if nil != err {
		return err
	}

	usm.hash = hash
	usm.auth_proto = auth_proto
	usm.auth_passphrase = value
	if 0 != int(hash) {
		usm.auth_key, err = generate_keys(hash, value)
		if nil != err {
			return errors.New("generate auth key failed - " + err.Error())
		}
	}

	priv_proto, value, err := getPriv(params)
	if nil != err {
		return err
	}

	usm.priv_proto = priv_proto
	usm.priv_passphrase = value

	if 0 != int(hash) {
		usm.priv_key, err = generate_keys(hash, value)
		if nil != err {
			return errors.New("generate priv key failed - " + err.Error())
		}
	}

	return nil
}

func (usm *USM) IsLocalize() bool {
	return nil != usm.localization_auth_key
}
func (usm *USM) Localize(key []byte) (err error) {
	if 0 == int(usm.hash) {
		return nil
	}

	usm.localization_auth_key, err = generate_localization_keys(usm.hash, usm.auth_key, key)
	if nil != err {
		return err
	}

	usm.localization_priv_key, err = generate_localization_keys(usm.hash, usm.priv_key, key)
	if nil != err {
		return err
	}
	return nil
}

//  typedef struct snmp_user {
//	enum snmp_authentication	auth_proto;
//	enum snmp_privacy       		priv_proto;
//	uint8_t                 				auth_key[SNMP_AUTH_KEY_SIZ];
//	size_t              auth_len;
//	uint8_t				priv_key[SNMP_PRIV_KEY_SIZ];
//	size_t              priv_len;
//	char				sec_name[SNMP_ADM_STR32_SIZ];
// } snmp_user_t;

func (usm *USM) Write(user *C.snmp_user_t) error {

	user.auth_proto = uint32(usm.auth_proto)
	user.priv_proto = uint32(usm.priv_proto)

	err := strcpy(&user.sec_name[0], SNMP_ADM_STR32_LEN, usm.name)
	if nil != err {
		return fmt.Errorf("sec_name too long")
	}

	err = memcpy(&user.auth_key[0], SNMP_AUTH_KEY_LEN, usm.localization_auth_key)
	if nil != err {
		return fmt.Errorf("auth_key too long")
	}
	user.auth_len = C.size_t(len(usm.localization_auth_key))

	err = memcpy(&user.priv_key[0], SNMP_AUTH_KEY_LEN, usm.localization_priv_key)
	if nil != err {
		return fmt.Errorf("priv_key too long")
	}
	user.priv_len = C.size_t(len(usm.localization_priv_key))
	return nil
}

func (usm *USM) Read(user *C.snmp_user_t) error {
	usm.auth_proto = AuthType(user.auth_proto)
	usm.priv_proto = PrivType(user.priv_proto)
	usm.name = readGoString(&user.sec_name[0], SNMP_ADM_STR32_LEN)
	usm.localization_auth_key = readGoBytes(&user.auth_key[0], C.uint32_t(user.auth_len))
	usm.localization_priv_key = readGoBytes(&user.priv_key[0], C.uint32_t(user.priv_len))
	return nil
}

func (usm *USM) String() string {
	if "" != usm.auth_passphrase {
		return fmt.Sprintf("auth = '[%s]%s' and priv = '[%s]%s'",
			usm.auth_proto.String(),
			usm.auth_key,
			usm.priv_proto.String(),
			usm.priv_key)
	}

	return fmt.Sprintf("auth = '[%s]%s' and priv = '[%s]%s'",
		usm.auth_proto.String(), hex.EncodeToString(usm.auth_key),
		usm.priv_proto.String(), hex.EncodeToString(usm.priv_key))
}
