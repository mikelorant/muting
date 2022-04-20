package certificates

import (
	"io/ioutil"
	"os"
	"fmt"
)

func WriteCertificates(filepath string, caConfig *CAConfig, serverConfig *ServerConfig) (err error) {
	if err = os.MkdirAll(filepath, 0o755); err != nil {
		return fmt.Errorf("WriteCertificates: unable to create certificates directory: %w", err)
	}

	if err = ioutil.WriteFile(filepath+"/ca.crt", caConfig.certificatePEM.Bytes(), 0o644); err != nil {
		return fmt.Errorf("WriteCertificates: unable to create CA certificate PEM file: %w", err)
	}

	if err = ioutil.WriteFile(filepath+"/tls.crt", serverConfig.certificatePEM.Bytes(), 0o644); err != nil {
		return fmt.Errorf("WriteCertificates: unable to create server certificate PEM file: %w", err)
	}

	if err = ioutil.WriteFile(filepath+"/tls.key", serverConfig.keyPEM.Bytes(), 0o644); err != nil {
		return fmt.Errorf("WriteCertificates: unable to create server key PEM file: %w", err)
	}

	return err
}
