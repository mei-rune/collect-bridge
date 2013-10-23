package license

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	Pwd                = []byte("89s7fd@f#gf0_h9)fg9g87s^d(6q3wke")
	TsnPublicKeyString = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnZVEnreOKLFUkNr6X9Ln
N/CMheb2rqfcfGhIMxs0/CsjpRq7X0nCANGnoFhlAqZs4wxdw3cSIpRvytFN5mNs
JrEjm00Al1tyu0a+/n6IVqGLlomt0YAAyKpdvzQ/TmBTS3NCSlkn2wBA2xolB/30
ZDmNnZDVFY/hrHCRrZLa6wTz5UopRXPavsnGLFXGjJLtJPPFeJEeYpDlJer3fYlq
AUUOjNI94aUftXwTipGU8aaprDkpKEmvgEMpATNhOCjwPXCV5hhF9znBl/YHCog6
t3uco9GZEMBbxj9wdjTAMeKzru2hkyN9p7XsbMf8cXs3DpVA2YtMMkqh52DZK1vZ
uQIDAQAB
-----END PUBLIC KEY-----`

	TsnPublicKey *rsa.PublicKey

	Title = ""
)

func init() {
	block, _ := pem.Decode([]byte(TsnPublicKeyString))
	if block == nil {
		panic("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err.Error())
	}
	TsnPublicKey = pubInterface.(*rsa.PublicKey)
}

// 加密
func Encrypt(origData []byte) (string, error) {
	encrypted_data, err := rsa.EncryptPKCS1v15(rand.Reader, TsnPublicKey, origData)
	if nil != err {
		return "", err
	}
	return hex.EncodeToString(encrypted_data), nil
}

// // 解密
// func RsaDecrypt(ciphertext string) ([]byte, error) {
// 	data, err := hex.DecodeString(ciphertext)
// 	if nil != err {
// 		return nil, err
// 	}

// 	block, _ := pem.Decode(privateKey)
// 	if block == nil {
// 		return nil, errors.New("private key error!")
// 	}
// 	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
// 	if nil != err {
// 		return nil, err
// 	}
// 	return rsa.DecryptPKCS1v15(rand.Reader, priv, data)
// }

func DecryptoFile(file string) (map[string]interface{}, error) {
	data, e := ioutil.ReadFile(file)
	if nil != e {
		return nil, errors.New("读 license 文件失败 -" + e.Error())
	}
	return DecryptoLicense(data)
}

func DecryptoLicense(data []byte) (map[string]interface{}, error) {
	encrypted_data, e := hex.DecodeString(string(data))
	if nil != e {
		return nil, errors.New("解密对象失败(第一步) -" + e.Error())
	}

	block, e := aes.NewCipher(Pwd)
	if nil != e {
		return nil, errors.New("创建加密对象失败 -" + e.Error())
	}
	dst := make([]byte, len(encrypted_data))
	for i := 0; ; i += block.BlockSize() {
		if (i + block.BlockSize()) >= len(encrypted_data) {
			block.Decrypt(dst[i:], encrypted_data[i:])
			break
		}
		block.Decrypt(dst[i:i+block.BlockSize()], encrypted_data[i:i+block.BlockSize()])
	}

	var attributes map[string]string
	if e = json.Unmarshal(dst, &attributes); nil != e {
		return nil, errors.New("解析数据失败(第一步) -" + e.Error())
	}

	sign_string, ok := attributes["sign"]
	if !ok {
		return nil, errors.New("没有找到答名！")
	}

	hex_data, ok := attributes["data"]
	if nil != e {
		return nil, errors.New("重编码数据失败 -" + e.Error())
	}
	if e = Verify([]byte(hex_data), sign_string); nil != e {
		return nil, errors.New("答名不正确 -" + e.Error())
	}

	json_data, e := hex.DecodeString(hex_data)
	if nil != e {
		return nil, errors.New("解密对象失败(第一步) -" + e.Error())
	}
	var properties map[string]interface{}
	decoder := json.NewDecoder(bytes.NewBuffer(json_data))
	decoder.UseNumber()
	e = decoder.Decode(&properties)
	if nil != e {
		return nil, errors.New("解析数据失败(第二步) -" + e.Error())
	}

	return properties, nil
}

func Verify(data []byte, sign_string string) error {
	sha := sha1.New()
	sha.Write(data)

	sign, err := hex.DecodeString(sign_string)
	if nil != err {
		return errors.New("sign is decode failed, " + err.Error())
	}

	err = rsa.VerifyPKCS1v15(TsnPublicKey, crypto.SHA1, sha.Sum(nil), sign)
	if nil != err {
		return errors.New("sign is verify failed, " + err.Error())
	}
	return nil
}

func abs(pa string) string {
	s, e := filepath.Abs(pa)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func SearchLicenseFile(nm string) (string, error) {
	files := []string{abs(nm),
		abs(filepath.Join("data", "tpt.lic")),
		abs(filepath.Join("..", "data", "tpt.lic"))}
	for _, file := range files {
		if st, e := os.Stat(file); nil == e && nil != st && !st.IsDir() {
			return file, nil
		}
	}
	return "", errors.New("license file is not exists, search path is:\r\n    " + strings.Join(files, "\r\n    "))
}
