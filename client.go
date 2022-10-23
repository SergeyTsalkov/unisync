package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/myssh/externalssh"
	"unisync/myssh/internalssh"
)

func runClient(conf *config.Config) {
	retryTime := 5 * time.Second

	for {
		err := _runClient(conf)
		if err != nil {
			log.Warnln("Client disconnected:", err)
		}

		log.Printf("Retrying in %v..", retryTime)
		time.Sleep(retryTime)
	}
}

func _runClient(conf *config.Config) error {
	var in io.Reader
	var out io.Writer
	var err error

	log.Printf("Connecting to %v@%v (%v)", conf.User, conf.Host, conf.Method)

	if conf.Method == "internalssh" {
		sshclient, err := internalssh.New(conf)
		if err != nil {
			return err
		}

		out, in, err = sshclient.Run()
		if err != nil {
			return fmt.Errorf("ssh error: %v", err)
		}

	} else if conf.Method == "ssh" {
		sshclient := externalssh.New(conf)
		out, in, err = sshclient.Run()
		if err != nil {
			return fmt.Errorf("ssh error: %v", err)
		}

	} else if conf.Method == "directtls" {
		cert, capool, err := getCert(false)
		if err != nil {
			return err
		}

		tlsdialer := &tls.Dialer{
			NetDialer: &net.Dialer{
				Timeout:   config.Duration(conf.ConnectTimeout),
				KeepAlive: config.Duration(conf.Timeout),
			},
			Config: &tls.Config{
				ServerName:   "unisync",
				Certificates: cert,
				RootCAs:      capool,
			},
		}

		conn, err := tlsdialer.Dial("tcp", fmt.Sprintf("%v:%v", conf.Host, conf.Port))
		if err != nil {
			return err
		}
		out = conn
		in = conn
	} else {
		panic("conf.Method=" + conf.Method)
	}

	c, err := client.New(in, out, conf)
	if err != nil {
		return err
	}
	return c.Run()
}
