package certificates

import (
	"io/ioutil"
	"os"
)

func WriteCertificates(filepath string, caConfig *CAConfig, serverConfig *ServerConfig) error {
	err := os.MkdirAll(filepath, 0o755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath+"/ca.crt", caConfig.certificatePEM.Bytes(), 0o644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath+"/tls.crt", serverConfig.certificatePEM.Bytes(), 0o644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath+"/tls.key", serverConfig.keyPEM.Bytes(), 0o644)
	if err != nil {
		return err
	}

	return nil
}
