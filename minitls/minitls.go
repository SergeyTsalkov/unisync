package minitls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
)

type MiniTLS struct {
	PrivateKey string // PEM
	privateKey *rsa.PrivateKey

	CaCert string // PEM
	caCert *x509.Certificate
}

func (m *MiniTLS) makePrivateKey() error {
	var err error
	m.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	encodedKey := &bytes.Buffer{}
	err = pem.Encode(encodedKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(m.privateKey),
	})
	if err != nil {
		return err
	}

	m.PrivateKey = string(encodedKey.Bytes())
	return nil
}

func (m *MiniTLS) makeCA() error {
	if m.privateKey == nil {
		return fmt.Errorf("Can't make CA: privateKey is not set")
	}

	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2019),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &m.privateKey.PublicKey, m.privateKey)
	if err != nil {
		return err
	}

	encodedCA := &bytes.Buffer{}
	err = pem.Encode(encodedCA, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return err
	}

	m.CaCert = string(encodedCA.Bytes())
	m.caCert, err = x509.ParseCertificate(caBytes)
	if err != nil {
		return err
	}

	return nil
}

func (m *MiniTLS) MakeCert() (*tls.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, m.caCert, &certPrivKey.PublicKey, m.privateKey)
	if err != nil {
		return nil, err
	}

	certPEM := &bytes.Buffer{}
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, err
	}

	certPrivKeyPEM := &bytes.Buffer{}
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		return nil, err
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}
	return &serverCert, nil
}

func New() (*MiniTLS, error) {
	mtls := &MiniTLS{}
	err := mtls.makePrivateKey()
	if err != nil {
		return nil, err
	}

	err = mtls.makeCA()
	if err != nil {
		return nil, err
	}

	return mtls, err
}
