package certificates

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	caConfig       *CAConfig
	commonName     string
	dnsNames       []string
	certificate    *x509.Certificate
	certificatePEM *bytes.Buffer
	key            *rsa.PrivateKey
	keyPEM         *bytes.Buffer
}

func NewServerCertificate(caConfig *CAConfig, commonName string, dnsNames []string) *ServerConfig {
	server := &ServerConfig{
		caConfig:   caConfig,
		commonName: commonName,
		dnsNames:   dnsNames,
	}
	server.genKey()
	server.genKeyPEM()
	server.genCertificate()
	server.genCertificatePEM()

	return server
}

func (s *ServerConfig) genCertificate() {
	certificate := &x509.Certificate{
		DNSNames:     s.dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   s.commonName,
			Organization: []string{"muting.io"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	s.certificate = certificate
}

func (s *ServerConfig) genKey() {
	key, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		log.Panic(err)
	}

	s.key = key
}

func (s *ServerConfig) genKeyPEM() {
	keyPEM := new(bytes.Buffer)
	_ = pem.Encode(keyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(s.key),
	})

	s.keyPEM = keyPEM
}

func (s *ServerConfig) genCertificatePEM() {
	certificate, err := x509.CreateCertificate(cryptorand.Reader, s.certificate, s.caConfig.certificate, &s.key.PublicKey, s.caConfig.key)
	if err != nil {
		log.Panic(err)
	}

	certificatePEM := new(bytes.Buffer)
	_ = pem.Encode(certificatePEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate,
	})

	s.certificatePEM = certificatePEM
}
