package tlsclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"unisync/config"
)

type tlsClient struct {
	dialer *tls.Dialer
	host   string
	conn   net.Conn
}

func New(conf *config.Config, cert []tls.Certificate, capool *x509.CertPool) *tlsClient {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout:   conf.ConnectTimeout,
			KeepAlive: conf.Timeout,
		},
		Config: &tls.Config{
			ServerName:   "unisync",
			Certificates: cert,
			RootCAs:      capool,
		},
	}

	return &tlsClient{
		dialer: dialer,
		host:   fmt.Sprintf("%v:%v", conf.Host, conf.Port),
	}

}

func (t *tlsClient) Run() (io.Writer, io.Reader, error) {
	var err error
	t.conn, err = t.dialer.Dial("tcp", t.host)
	return t.conn, t.conn, err
}

func (t *tlsClient) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
