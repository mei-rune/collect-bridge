package snmp

// #include "bsnmp/config.h"
// #include <stdlib.h>
// #include "bsnmp/asn1.h"
// #include "bsnmp/snmp.h"
import "C"

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

type securityModelWithCopy interface {
	SecurityModel
	Write(*C.snmp_user_t) error
	Read(*C.snmp_user_t) error
}

type HashUSM struct {
	auth_proto AuthType
	priv_proto PrivType
	auth_key   []byte
	priv_key   []byte

	auth_key_with_engine_id []byte
	priv_key_with_engine_id []byte
	name                    string
}

func getAuth(params map[string]string) (AuthType, string, error) {
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

func (usm *HashUSM) Write(user *C.snmp_user_t) error {
	//  typedef struct snmp_user {
	//	enum snmp_authentication	auth_proto;
	//	enum snmp_privacy       		priv_proto;
	//	uint8_t                 				auth_key[SNMP_AUTH_KEY_SIZ];
	//	size_t              auth_len;
	//	uint8_t				priv_key[SNMP_PRIV_KEY_SIZ];
	//	size_t              priv_len;
	//	char				sec_name[SNMP_ADM_STR32_SIZ];
	// } snmp_user_t;

	user.auth_proto = uint32(usm.auth_proto)
	user.priv_proto = uint32(usm.priv_proto)

	err := strcpy(&user.sec_name[0], SNMP_ADM_STR32_LEN, usm.name)
	if nil != err {
		return fmt.Errorf("sec_name too long")
	}

	err = memcpy(&user.auth_key[0], SNMP_AUTH_KEY_LEN, usm.auth_key)
	if nil != err {
		return fmt.Errorf("auth_key too long")
	}
	user.auth_len = C.size_t(len(usm.auth_key))

	err = memcpy(&user.priv_key[0], SNMP_AUTH_KEY_LEN, usm.priv_key)
	if nil != err {
		return fmt.Errorf("priv_key too long")
	}
	user.priv_len = C.size_t(len(usm.priv_key))
	return nil
}

func (usm *HashUSM) Read(user *C.snmp_user_t) error {
	//  typedef struct snmp_user {
	//	enum snmp_authentication	auth_proto;
	//	enum snmp_privacy       		priv_proto;
	//	uint8_t                 				auth_key[SNMP_AUTH_KEY_SIZ];
	//	size_t              auth_len;
	//	uint8_t				priv_key[SNMP_PRIV_KEY_SIZ];
	//	size_t              priv_len;
	//	char				sec_name[SNMP_ADM_STR32_SIZ];
	// } snmp_user_t;

	usm.auth_proto = AuthType(user.auth_proto)
	usm.priv_proto = PrivType(user.priv_proto)
	usm.name = readGoString(&user.sec_name[0], SNMP_ADM_STR32_LEN)
	usm.auth_key = readGoBytes(&user.auth_key[0], C.uint32_t(user.auth_len))
	usm.priv_key = readGoBytes(&user.priv_key[0], C.uint32_t(user.priv_len))
	return nil
}

func (usm *HashUSM) Init(params map[string]string) error {
	name, ok := params["secname"]
	if !ok {
		return errors.New("secname is required.")
	}
	usm.name = name

	auth_proto, value, err := getAuth(params)
	if nil != err {
		return err
	}

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

func (usm *HashUSM) String() string {
	return fmt.Sprintf("auth = '[%s]%s' and priv = '[%s]%s'",
		usm.auth_proto.String(), hex.EncodeToString(usm.auth_key),
		usm.priv_proto.String(), hex.EncodeToString(usm.priv_key))
}

type USM struct {
	auth_proto AuthType
	priv_proto PrivType
	auth_key   string
	priv_key   string
	name       string
}

func (usm *USM) Write(user *C.snmp_user_t) error {
	//  typedef struct snmp_user {
	//	enum snmp_authentication	auth_proto;
	//	enum snmp_privacy       		priv_proto;
	//	uint8_t                 				auth_key[SNMP_AUTH_KEY_SIZ];
	//	size_t              auth_len;
	//	uint8_t				priv_key[SNMP_PRIV_KEY_SIZ];
	//	size_t              priv_len;
	//	char				sec_name[SNMP_ADM_STR32_SIZ];
	// } snmp_user_t;

	user.auth_proto = uint32(usm.auth_proto)
	user.priv_proto = uint32(usm.priv_proto)

	err := strcpy(&user.sec_name[0], SNMP_ADM_STR32_LEN, usm.name)
	if nil != err {
		return fmt.Errorf("sec_name too long")
	}

	s := C.CString(usm.auth_key)
	defer func() {
		if nil != s {
			C.free(unsafe.Pointer(s))
		}
	}()
	ret_code := C.snmp_set_auth_passphrase(user, s, C.size_t(len(usm.auth_key)))
	if 0 != ret_code {
		return errors.New("set auth key failed - " + C.GoString(C.snmp_get_error(ret_code)))
	}
	C.free(unsafe.Pointer(s))
	s = nil

	s = C.CString(usm.priv_key)
	ret_code = C.snmp_set_priv_passphrase(user, s, C.size_t(len(usm.priv_key)))
	if 0 != ret_code {
		return errors.New("set priv key failed - " + C.GoString(C.snmp_get_error(ret_code)))
	}
	return nil
}

func (usm *USM) Read(user *C.snmp_user_t) error {
	return errors.New("not implemented")
}

func (usm *USM) Init(params map[string]string) error {
	name, ok := params["secname"]
	if !ok {
		return errors.New("secname is required.")
	}
	usm.name = name

	auth_proto, value, err := getAuth(params)
	if nil != err {
		return err
	}

	usm.auth_proto = auth_proto
	usm.auth_key = value

	priv_proto, value, err := getPriv(params)
	if nil != err {
		return err
	}

	usm.priv_proto = priv_proto
	usm.priv_key = value
	return nil
}

func (usm *USM) String() string {
	return fmt.Sprintf("auth = '[%s]%s' and priv = '[%s]%s'",
		usm.auth_proto.String(),
		usm.auth_key,
		usm.priv_proto.String(),
		usm.priv_key)
}
