package minica

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

var keySize = 2048

type MiniCA struct {
	privateKey *rsa.PrivateKey
	caCert     *x509.Certificate
	serverCert []tls.Certificate
}

func (m *MiniCA) makeCA() error {
	if m.privateKey == nil {
		return fmt.Errorf("Can't make CA: privateKey is not set")
	}

	ca := &x509.Certificate{
		NotBefore:             time.Now().AddDate(-10, 0, 0),
		NotAfter:              time.Now().AddDate(10, 0, 0),
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

	m.caCert, err = x509.ParseCertificate(caBytes)
	if err != nil {
		return err
	}

	return nil
}

func (m *MiniCA) GetCAPool() *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(m.caCert)
	return pool
}

func (m *MiniCA) GetCert() ([]tls.Certificate, error) {
	if m.serverCert == nil {
		var err error
		m.serverCert, err = m.makeCert()
		if err != nil {
			return nil, err
		}
	}
	return m.serverCert, nil
}

func (m *MiniCA) makeCert() ([]tls.Certificate, error) {
	cert := &x509.Certificate{
		NotBefore:    time.Now().AddDate(-10, 0, 0),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		DNSNames:     []string{"unisync"},
		SerialNumber: big.NewInt(2019),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, keySize)
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
	return []tls.Certificate{serverCert}, nil
}

func New(fullpath string) (*MiniCA, error) {
	m := &MiniCA{}
	var err error

	m.privateKey, err = rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	if err := m.makeCA(); err != nil {
		return nil, err
	}
	if err := m.save(fullpath); err != nil {
		return nil, err
	}

	return m, nil
}
