package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"io"
	"log"
	"os"
)

var (
	bits         = flag.Int("bits", 2048, "密钥长度，默认为2048位")
	rand_file    = flag.String("rand", "", "随机数生成文件")
	private_file = flag.String("private", "private.pem", "私有密钥文件输出路径")
	public_file  = flag.String("public", "public.pem", "公有密钥文件输出路径")
)

func main() {
	flag.Parse()

	if 0 == len(*rand_file) {
		log.Fatal("随机数生成文件不能为空！")
		return
	}

	f, e := os.Open(*rand_file)
	if nil != e {
		log.Fatal("打开随机数生成文件失败 -", e)
		return
	}
	defer f.Close()

	if err := GenRsaKey(f, *bits); err != nil {
		log.Fatal("密钥文件生成失败 -", err)
		return
	}

	log.Println("密钥文件生成成功！")
}

func GenRsaKey(reader io.Reader, bits int) error {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(reader, bits)
	if err != nil {
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	file, err := os.Create(*private_file)
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	// 生成公钥文件
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	file, err = os.Create(*public_file)
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}
