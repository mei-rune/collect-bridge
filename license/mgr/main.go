package main

import (
	"crypto"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"license"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	cmd                 = flag.String("cmd", "", "create, import")
	file                = flag.String("file", "", "数据文件")
	licensePath         = flag.String("lic", "", "License文件")
	TsnPrivateKeyString = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAnZVEnreOKLFUkNr6X9LnN/CMheb2rqfcfGhIMxs0/CsjpRq7
X0nCANGnoFhlAqZs4wxdw3cSIpRvytFN5mNsJrEjm00Al1tyu0a+/n6IVqGLlomt
0YAAyKpdvzQ/TmBTS3NCSlkn2wBA2xolB/30ZDmNnZDVFY/hrHCRrZLa6wTz5Uop
RXPavsnGLFXGjJLtJPPFeJEeYpDlJer3fYlqAUUOjNI94aUftXwTipGU8aaprDkp
KEmvgEMpATNhOCjwPXCV5hhF9znBl/YHCog6t3uco9GZEMBbxj9wdjTAMeKzru2h
kyN9p7XsbMf8cXs3DpVA2YtMMkqh52DZK1vZuQIDAQABAoIBAC4feHw0IYnLjYLw
dQQDCOYYpCi1F1K7kw9evnMm7XU5cy9qCZm0TvJKaxPFi5sg9xHllrQVb9trMuVc
Kb7bLtaMVm2oNhoXDBfAdzqp8mHY2rBvD88X9iLFqrbCJh1cmESnMantOnshMdpv
ZpNWQ2fqaIbL03KCMH12XU0+hJDw5paLdUVEp4gmdWquwYPoeBoY7ls0gZf38x9/
A4sJrj8/AgLU6Zz4AMy9T2vUuDnrwi8HBszPhmaIjuEFJY5t6J7aNlcMRufuUeB2
NhLXA4LtU1cegn5cgaq+fXbQbMIO/xPL8x2gZbYRfUaGNbNnglFQBltspu2ZSoZP
i5C+V3kCgYEAyEgrRCQgSIkEJLgCAAAAiUQkCOhomAEASIPEYMMAAABIg+wwSI1M
JDhIiQwk6E62AQBIiUQkKEiLRCQ4SIkEJEiLRCRASIlEJAhIjUQkSEiJRCQQSMdE
JBgAAAAASItEJChIiUQkIOi09P//SIPEMMMAAAAAAAAAAAAAAAAAAAMCgYEAyWwk
DItLIDnNdQYx7YlsJAz/xotDIDnGD4xr////uKANXgBIiQQk6PSDAACLbCQMSItc
JBjpZv///wAAAAAAAGVIiwwlKAAAAEiLCUg7IXcF6DpVAQBIg+wQTItEJCBIi3wk
MEiLbCQYi0Ugg/gAdQcxwEiDxBDDi00gSGPJSJMCgYASL7N2EY246HAu4WKVG2rx
C/X1tRziSJz8+LIZUzupxFzRVd4giGwUkePMRgUH7zXJq3vqsvxRiBzVSDCwLXjp
zoiO3HfV3lkIqJPl9/0Pbz6/qEKuSSHf4SoG1fkwnSzH17yWclCRiG/+G0zUCdsD
zoEuftGBLn7RgS5+0YEufwKBgA3VXIB0DJvQ/HheDY3H02KCdgqXw+txrFWjCkPR
vClc7K4ZDOJEI5jQOjME1o59x04wLAranyUQze8ge9+EIHvmeN373o1pupZKZSol
CNKajxBM+UuRTmmpC9GF+w8UHIdf5kheoxyxOfJOtBX8MWPYM6WGFW6gzB1EFcSC
/e77AoGAMVAuH9jhcFFw2wMQBLVgjHsylOpN0QxzI2j5TJDXiMYQqaYLjZfwTSEH
ktQSwnhzGU1DrV72FfPFmHbuTkIoOT6Zl4d5fF+ZpsQswt1mkiC79RaGkTUhSP2a
b7tN1CKz5YFlbm5qIV6SSlyiPD3NCKmE5Aekh8armSqAGnQGg2U=
-----END RSA PRIVATE KEY-----`

	TsnPrivateKey *rsa.PrivateKey
)

func init() {
	block, _ := pem.Decode([]byte(TsnPrivateKeyString))
	if block == nil {
		panic("private key error")
	}
	privInterface, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err.Error())
	}
	TsnPrivateKey = privInterface
}

func main() {
	flag.Parse()
	args := flag.Args()
	if nil != args {
		switch len(args) {
		case 1:
			flag.Set("cmd", args[0])
		case 2:
			flag.Set("cmd", args[0])
			flag.Set("file", args[1])
		default:
			fmt.Println("arguments is too much.")
			os.Exit(-1)
			return
		}
	}

	switch *cmd {
	case "open":
		open_cmd()
	case "sign":
		gen_cmd()
	default:
		fmt.Println("unknown command -", *cmd)
		return
	}
}

func gen_cmd() {
	if 0 == len(*file) {
		log.Fatal("数据文件不能为空！")
		return
	}

	data, e := ioutil.ReadFile(*file)
	if nil != e {
		log.Fatal("读数据文件失败 -", e)
		return
	}

	if nil == data || 0 == len(data) {
		log.Fatal("读数据文件失败 - 内容为空")
		return
	}

	hex_data := hex.EncodeToString(data)
	sign, e := Sign([]byte(hex_data))
	if nil != e {
		log.Fatal("生成 license 文件失败 -", e)
		return
	}

	if 0 == len(*licensePath) {
		pa := filepath.Dir(*file)
		nm := filepath.Join(pa, strings.TrimSuffix(filepath.Base(*file), ".tsn")+".lic")
		flag.Set("lic", nm)
	}

	attributes := map[string]string{"sign": sign, "data": hex_data}
	json_data, e := json.Marshal(attributes)
	if nil != e {
		log.Fatal("重编码数据失败 --", e)
		return
	}

	block, e := aes.NewCipher(license.Pwd)
	if nil != e {
		log.Fatal("创建加密对象失败 -", e)
		return
	}
	data_length := (len(json_data)/block.BlockSize() + 1) * block.BlockSize()
	for i := len(json_data); i < data_length; i++ {
		json_data = append(json_data, ' ')
	}
	dst := make([]byte, len(json_data))
	for i := 0; ; i += block.BlockSize() {
		if (i + block.BlockSize()) >= len(json_data) {
			block.Encrypt(dst[i:], json_data[i:])
			break
		}
		block.Encrypt(dst[i:i+block.BlockSize()], json_data[i:i+block.BlockSize()])
	}

	e = ioutil.WriteFile(*licensePath, []byte(hex.EncodeToString(dst)), 0)
	if nil != e {
		log.Fatal("写 license 文件失败 -", e)
		return
	}
}

func open_cmd() {
	if 0 == len(*file) {
		log.Fatal("数据文件不能为空！")
		return
	}

	data, e := ioutil.ReadFile(*file)
	if nil != e {
		log.Fatal("读数据文件失败 -", e)
		return
	}

	if nil == data || 0 == len(data) {
		log.Fatal("读数据文件失败 - 内容为空")
		return
	}
	origin_data, e := Decrypt(string(data))
	if nil != e {
		log.Fatal("解密文件失败 -", e)
		return
	}

	fmt.Println(string(origin_data))
}

func Decrypt(encrypted_data string) ([]byte, error) {
	data, e := hex.DecodeString(encrypted_data)
	if nil != e {
		return nil, e
	}

	return rsa.DecryptPKCS1v15(rand.Reader, TsnPrivateKey, data)
}

func Sign(data []byte) (string, error) {
	sha := sha1.New()
	sha.Write([]byte(data))
	s, err := rsa.SignPKCS1v15(rand.Reader, TsnPrivateKey, crypto.SHA1, sha.Sum(nil))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(s), nil
}
