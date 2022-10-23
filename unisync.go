package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unisync/config"
	"unisync/log"
	"unisync/minica"
)

var mca *minica.MiniCA

func main() {
	debugFlag := flag.Bool("debug", false, "debug mode")
	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	serverFlag := flag.String("server", "", "run server")
	flag.Parse()
	args := flag.Args()
	var conf *config.Config

	if len(args) == 1 {
		var err error
		conf, err = config.Parse(args[0])
		if err != nil {
			log.Fatalln(err)
		}
	} else if len(args) == 2 {
		userhost, remotepath, valid := strings.Cut(args[1], ":")
		if !valid {
			showHelp()
		}

		user, host, valid := strings.Cut(userhost, "@")
		if !valid {
			showHelp()
		}

		conf = config.New()
		conf.Local = args[0]
		conf.Remote = remotepath
		conf.User = user
		conf.Host = host
	}

	if *debugFlag {
		conf.Debug = true
	}
	if conf.Debug {
		log.Reset()
		log.Add(os.Stdout, log.Debug, "")
	}

	if *stdServerFlag {
		err := runStdinServer()
		if err != nil {
			log.Fatalln(err)
		}

	} else if *serverFlag != "" {
		err := runDirectServer(*serverFlag)
		if err != nil {
			log.Fatalln(err)
		}

	} else {
		if conf == nil {
			showHelp()
		}

		runClient(conf)
	}
}

func getCert(canMake bool) ([]tls.Certificate, *x509.CertPool, error) {
	if mca == nil {
		var err error
		fullpath := filepath.Join(config.ConfigDir(), "secure.key")
		mca, err = minica.Load(fullpath)

		if err != nil && canMake && errors.Is(err, fs.ErrNotExist) {
			mca, err = minica.New(fullpath)
			if err != nil {
				return nil, nil, fmt.Errorf("Failed to create key at %v: %w", fullpath, err)
			}

			log.Printf("Created new key at %v, make sure to copy this to the client so it can connect!", fullpath)
		} else if err != nil {
			return nil, nil, fmt.Errorf("Failed to load key at %v: %w", fullpath, err)
		}
	}

	cert, err := mca.GetCert()
	if err != nil {
		return nil, nil, err
	}

	return cert, mca.GetCAPool(), nil
}
