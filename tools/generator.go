package tools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

type RSAGenarator struct{}

func (this *RSAGenarator) Generate(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return privateKey, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	file, err := os.Create("privatekey.pem")
	if err != nil {
		return privateKey, err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return privateKey, err
	}

	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return privateKey, err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	file, err = os.Create("publickey.pem")
	if err != nil {
		return privateKey, err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return privateKey, err
	}
	return privateKey, nil
}
