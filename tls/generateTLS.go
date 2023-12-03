package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"github.com/pion/dtls/v2"
)
func GenerateDTLSConfig() (*dtls.Config, error) {
	// 生成RSA密钥对
	key, err := rsa.GenerateKey(rand.Reader, 2048) // 使用2048位密钥长度更安全
	if err != nil {
		return nil, err
	}

	// 创建一个自签名证书
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		// 可以在这里添加更多证书字段
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	// 将密钥和证书编码为PEM格式
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// 使用PEM编码的密钥和证书创建tls.Certificate
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	// 创建并返回DTLS配置
	return &dtls.Config{
		Certificates: []tls.Certificate{tlsCert},
		// 这里可以添加更多配置项，例如CipherSuites, SRTPProtectionProfiles等
	}, nil
}
func GenerateTLSConfig(nextProtos string) *tls.Config {
	// 生成RSA密钥对
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	// 创建一个自签名证书
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	// 将密钥和证书编码为PEM格式
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	// 返回包含TLS证书的TLS配置，并指定应用协议名称
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{nextProtos},
	}
}
