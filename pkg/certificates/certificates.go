package certificates

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type CAConfig struct {
	CA        *x509.Certificate
	CAPrivKey *rsa.PrivateKey
	CAPEM     *bytes.Buffer
}

type ServerConfig struct {
	ServerCertPEM    *bytes.Buffer
	ServerPrivKeyPEM *bytes.Buffer
}

func GenerateCA() *CAConfig {
	log.Info("Generating certificate authority.")

	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"velotio.com"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		log.Panic(err)
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Panic(err)
	}

	// PEM encode CA cert
	caPEM := new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	return &CAConfig{
		CA:        ca,
		CAPrivKey: caPrivKey,
		CAPEM:     caPEM,
	}
}

func GenerateServerCertificate(commonName string, dnsNames []string, caConfig *CAConfig) *ServerConfig {
	log.Info("Generating server certificates.")

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"velotio.com"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		log.Panic(err)
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, caConfig.CA, &serverPrivKey.PublicKey, caConfig.CAPrivKey)
	if err != nil {
		log.Panic(err)
	}

	// PEM encode the  server cert and key
	serverCertPEM := new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	return &ServerConfig{
		ServerCertPEM:    serverCertPEM,
		ServerPrivKeyPEM: serverPrivKeyPEM,
	}
}

func WriteCertificates(filepath string, caConfig *CAConfig, serverConfig *ServerConfig) error {
	log.Info(fmt.Sprintf("Writing certificates to: %s", filepath))

	err := os.MkdirAll(filepath, 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath+"/ca.crt", caConfig.CAPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath+"/tls.crt", serverConfig.ServerCertPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath+"/tls.key", serverConfig.ServerPrivKeyPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
