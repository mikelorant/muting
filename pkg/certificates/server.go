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

type ServerConfig struct {
	caConfig       *CAConfig
	commonName     string
	dnsNames       []string
	certificate    *x509.Certificate
	certificatePEM *bytes.Buffer
	key            *rsa.PrivateKey
	keyPEM         *bytes.Buffer
}

func NewServerCertificate(caConfig *CAConfig, commonName string, dnsNames []string) (server *ServerConfig, err error) {
	server = &ServerConfig{
		caConfig:   caConfig,
		commonName: commonName,
		dnsNames:   dnsNames,
	}
	if err = server.genKey(); err != nil {
		return nil, fmt.Errorf("NewServerCertificate: genKey failed: %w", err)
	}

	if err = server.genKeyPEM(); err != nil {
		return nil, fmt.Errorf("NewServerCertificate: genKeyPEM failed: %w", err)
	}

	server.genCertificate()

	if err = server.genCertificatePEM(); err != nil {
		return nil, fmt.Errorf("NewServerCertificate: genCertificatePEM failed: %w", err)
	}

	return server, err
}

func (s *ServerConfig) genKey() (err error) {
	s.key, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("genKey: unable to generate key: %w", err)
	}

	return err
}

func (s *ServerConfig) genKeyPEM() (err error) {
	err = pem.Encode(s.keyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(s.key),
	})
	if err != nil {
		return fmt.Errorf("genKeyPEM: unable to create key PEM: %w", err)
	}

	return err
}

func (s *ServerConfig) genCertificate() {
	s.certificate = &x509.Certificate{
		DNSNames:       s.dnsNames,
		SerialNumber:   big.NewInt(1658),
		Subject:        pkix.Name{
			CommonName:   s.commonName,
			Organization: []string{"muting.io"},
		},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().AddDate(1, 0, 0),
		SubjectKeyId:   []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:    []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:       x509.KeyUsageDigitalSignature,
	}
}

func (s *ServerConfig) genCertificatePEM() (err error) {
	var cert []byte

	cert, err = x509.CreateCertificate(cryptorand.Reader, s.certificate, s.caConfig.certificate, &s.key.PublicKey, s.caConfig.key)
	if err != nil {
		return fmt.Errorf("genCertificatePEM: unable to create certificate PEM: %w", err)
	}

	err = pem.Encode(s.certificatePEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	if err != nil {
		return fmt.Errorf("genCertificatePEM: unable to PEM encode certificate: %w", err)
	}

	return err
}
