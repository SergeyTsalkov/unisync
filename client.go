package main

import (
	"fmt"
	"io"
	"time"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/transports/externalssh"
	"unisync/transports/internalssh"
	"unisync/transports/tlsclient"
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
		defer sshclient.Close()

		out, in, err = sshclient.Run()
		if err != nil {
			return fmt.Errorf("ssh error: %v", err)
		}

	} else if conf.Method == "ssh" {

		sshclient := externalssh.New(conf)
		defer func() {
			err := sshclient.Close()
			if err != nil {
				log.Warnln("ssh client exited:", err)
			}
		}()
		out, in, err = sshclient.Run()
		if err != nil {
			return fmt.Errorf("ssh error: %v", err)
		}

	} else if conf.Method == "directtls" {

		cert, capool, err := getCert(false)
		if err != nil {
			return err
		}
		tlsc := tlsclient.New(conf, cert, capool)
		defer tlsc.Close()

		out, in, err = tlsc.Run()
		if err != nil {
			return err
		}

	} else {
		panic("conf.Method=" + conf.Method)
	}

	c, err := client.New(in, out, conf)
	if err != nil {
		return err
	}
	return c.Run()
}
