package minica

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func (m *MiniCA) save(fullpath string) error {
	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(m.privateKey),
	})
	if err != nil {
		return err
	}

	err = pem.Encode(file, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: m.caCert.Raw,
	})
	if err != nil {
		return err
	}

	return nil
}

func Load(fullpath string) (*MiniCA, error) {
	bytes, err := os.ReadFile(fullpath)
	if err != nil {
		return nil, err
	}

	m := &MiniCA{}
	for {
		var err error
		var p *pem.Block
		p, bytes = pem.Decode(bytes)
		if p == nil {
			break
		}

		if p.Type == "RSA PRIVATE KEY" {
			m.privateKey, err = x509.ParsePKCS1PrivateKey(p.Bytes)
			if err != nil {
				return nil, err
			}
		}
		if p.Type == "CERTIFICATE" {
			m.caCert, err = x509.ParseCertificate(p.Bytes)
			if err != nil {
				return nil, err
			}
		}
	}

	if m.privateKey == nil {
		return nil, fmt.Errorf("missing private key")
	}
	if m.caCert == nil {
		return nil, fmt.Errorf("missing certificate")
	}
	return m, nil
}
