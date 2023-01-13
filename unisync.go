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
	"runtime/debug"
	"unisync/background"
	"unisync/config"
	"unisync/log"
	"unisync/minica"
	"unisync/watcher"
)

var mca *minica.MiniCA

func main() {
	startFlag := flag.Bool("start", false, "start in background mode")
	stopFlag := flag.Bool("stop", false, "stop in background mode")
	stopAllFlag := flag.Bool("stopall", false, "stop all background instances")
	statusFlag := flag.Bool("status", false, "list instances running in background mode")

	versionFlag := flag.Bool("version", false, "show version and exit")
	debugFlag := flag.Bool("debug", false, "debug mode")
	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	serverFlag := flag.String("server", "", "run server")
	flag.Parse()
	args := flag.Args()
	var conf *config.Config

	if *versionFlag {
		fmt.Println("git revision:", gitRevision())
		fmt.Println("watcher:", watcher.Strategy)
		os.Exit(0)
	}

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
	if *stdServerFlag {
		err := runStdinServer()
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}
	if *serverFlag != "" {
		err := runDirectServer(*serverFlag)
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
	} else {
		showHelp()
	}

	if background.IsChild() {
		err := background.WritePid(conf.Name)
		if err != nil {
			log.Warnln("Error writing pid file:", err)
		}
	}

	if *startFlag && !background.IsChild() {
		err := background.Start(conf.Name)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)

	} else if *stopFlag && !background.IsChild() {
		err := background.Stop(conf.Name)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)

	} else {

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

func getCert(keyPath string, canMake bool) ([]tls.Certificate, *x509.CertPool, error) {
	if mca == nil {
		var err error

		if !filepath.IsAbs(keyPath) {
			keyPath = filepath.Join(config.ConfigDir(), keyPath)
		}

		mca, err = minica.Load(keyPath)

		if err != nil && canMake && errors.Is(err, fs.ErrNotExist) {
			mca, err = minica.New(keyPath)
			if err != nil {
				return nil, nil, fmt.Errorf("Failed to create key at %v: %w", keyPath, err)
			}

			log.Printf("Created new key at %v, make sure to copy this to the client so it can connect!", keyPath)
		} else if err != nil {
			return nil, nil, fmt.Errorf("Failed to load key at %v: %w", keyPath, err)
		}
	}

	cert, err := mca.GetCert()
	if err != nil {
		return nil, nil, err
	}

	return cert, mca.GetCAPool(), nil
}

func gitRevision() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}
