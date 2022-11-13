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
	"unisync/background"
	"unisync/config"
	"unisync/log"
	"unisync/minica"
)

var mca *minica.MiniCA

func main() {
	startFlag := flag.Bool("start", false, "start in background mode")
	stopFlag := flag.Bool("stop", false, "stop in background mode")
	stopAllFlag := flag.Bool("stopall", false, "stop all background instances")
	statusFlag := flag.Bool("status", false, "list instances running in background mode")

	debugFlag := flag.Bool("debug", false, "debug mode")
	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	serverFlag := flag.String("server", "", "run server")
	flag.Parse()
	args := flag.Args()
	var conf *config.Config

	if *statusFlag {
		running := background.ListRunning()

		if len(running) == 0 {
			fmt.Println("There are no background instances running.")
		} else {
			fmt.Println("Background instances running:")
			for _, name := range running {
				fmt.Println(name)
			}
		}

		os.Exit(0)
	}
	if *stopAllFlag {
		err := background.StopAll()
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}

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

		conf = config.New("")
		conf.Local = args[0]
		conf.Remote = remotepath
		conf.User = user
		conf.Host = host
		err := conf.Validate()
		if err != nil {
			log.Fatalln(err)
		}
	}

	if conf != nil && conf.Name != "" && background.IsChild() {
		err := background.WritePid(conf.Name)
		if err != nil {
			log.Warnln("Error writing pid file:", err)
		}
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

	} else if *startFlag && conf.Name != "" && !background.IsChild() {
		err := background.Start(conf.Name)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)

	} else if *stopFlag && conf.Name != "" && !background.IsChild() {
		err := background.Stop(conf.Name)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)

	} else {
		if conf == nil {
			showHelp()
		}

		if *debugFlag {
			conf.Debug = true
		}
		if conf.Debug {
			log.ScreenLevel = log.Debug
		}
		if conf.Log != "" {
			err := log.AddFile(conf.Log, log.ScreenLevel, "2006-01-02 15:04:05")
			if err != nil {
				log.Fatalf("Unable to open file for logging: %v", err)
			}
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
