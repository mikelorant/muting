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

type CAConfig struct {
	certificate    *x509.Certificate
	certificatePEM *bytes.Buffer
	key            *rsa.PrivateKey
}

func NewCACertificate() *CAConfig {
	ca := &CAConfig{}
	ca.genKey()
	ca.genCertificate()
	ca.genCertificatePEM()

	return ca
}

func (c *CAConfig) GetCertificatePEM() *bytes.Buffer {
	return c.certificatePEM
}

func (c *CAConfig) genKey() {
	key, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		log.Panic(err)
	}

	c.key = key
}

func (c *CAConfig) genCertificate() {
	certificate := &x509.Certificate{
		SerialNumber:          big.NewInt(2020),
		Subject:               pkix.Name{
			Organization: []string{"muting.io"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	c.certificate = certificate
}

func (c *CAConfig) genCertificatePEM() {
	certificate, err := x509.CreateCertificate(cryptorand.Reader, c.certificate, c.certificate, &c.key.PublicKey, c.key)
	if err != nil {
		log.Panic(err)
	}

	certificatePEM := new(bytes.Buffer)
	_ = pem.Encode(certificatePEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate,
	})

	c.certificatePEM = certificatePEM
}
