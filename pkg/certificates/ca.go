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
	"fmt"
)

type CAConfig struct {
	certificate    *x509.Certificate
	certificatePEM *bytes.Buffer
	key            *rsa.PrivateKey
}

func NewCACertificate() (ca *CAConfig, err error) {
	if err = ca.genKey(); err != nil {
		return nil, fmt.Errorf("NewCACertificate: genKey failed: %w", err)
	}

	ca.genCertificate()

	if err = ca.genCertificatePEM(); err != nil {
		return nil, fmt.Errorf("NewCACertificate: genCertificate failed: %w", err)
	}

	return ca, err
}

func (c *CAConfig) GetCertificatePEM() *bytes.Buffer {
	return c.certificatePEM
}

func (c *CAConfig) genKey() (err error) {
	c.key, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("genKey: unable to generate key: %w", err)
	}

	return err
}

func (c *CAConfig) genCertificate() {
	c.certificate = &x509.Certificate{
		SerialNumber:          big.NewInt(2020),
		Subject:               pkix.Name{
			Organization:        []string{"muting.io"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

func (c *CAConfig) genCertificatePEM() (err error) {
	var cert []byte

	cert, err = x509.CreateCertificate(cryptorand.Reader, c.certificate, c.certificate, &c.key.PublicKey, c.key)
	if err != nil {
		return fmt.Errorf("genCertificatePEM: unable to create certificate PEM: %w", err)
	}

	err = pem.Encode(c.certificatePEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	if err != nil {
		return fmt.Errorf("genCertificatePEM: unable to PEM encode certificate: %w", err)
	}

	return err
}
