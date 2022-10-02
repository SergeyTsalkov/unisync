package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Host   string `json:"host"`
	Method string `json:"method"`
	Prefer string `json:"prefer'`

	Chmod *configTypeChmod `json:"chmod"`
}

type configTypeChmod struct {
	Local     *fs.FileMode `json:"local"`
	LocalDir  *fs.FileMode `json:"local_dir"`
	Remote    *fs.FileMode `json:"remote"`
	RemoteDir *fs.FileMode `json:"remote_dir"`
	Mask      *fs.FileMode `json:"mask"`
	DirMask   *fs.FileMode `json:"dir_mask`
}

var once sync.Once
var configDir string

func Parse(path string) (*Config, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(ConfigDir(), path)
	}

	if !fileExists(path) {
		if !strings.HasSuffix(path, ".json") {
			path = path + ".json"
		}

		if !fileExists(path) {
			return nil, fmt.Errorf("ConfigFile %v does not exist", path)
		}
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read ConfigFile %v: %v", path, err)
	}

	config := &Config{}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse ConfigFile %v: %v", path, err)
	}

	config.Validate()
	return config, nil
}

func (C *Config) Validate() {
	if C.Chmod == nil {
		C.Chmod = &configTypeChmod{}
	}
	if C.Method == "" {
		C.Method = "ssh"
	}

	if C.Prefer == "" {
		C.Prefer = "newest"
	}
	if C.Prefer != "newest" && C.Prefer != "oldest" && C.Prefer != "local" && C.Prefer != "remote" {
		log.Fatalln("config.prefer must be one of: newest, oldest, local, remote")
	}

	if C.Chmod.Local == nil {
		C.Chmod.Local = makeMode(0644)
	}
	if C.Chmod.LocalDir == nil {
		C.Chmod.LocalDir = makeMode(0755)
	}
	if C.Chmod.Remote == nil {
		C.Chmod.Remote = makeMode(0644)
	}
	if C.Chmod.RemoteDir == nil {
		C.Chmod.RemoteDir = makeMode(0755)
	}
	if C.Chmod.Mask == nil {
		C.Chmod.Mask = makeMode(0100)
	}
	if C.Chmod.DirMask == nil {
		C.Chmod.DirMask = makeMode(0)
	}
}

func makeMode(mode fs.FileMode) *fs.FileMode {
	return &mode
}

func ConfigDir() string {
	once.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			log.Fatalln("Unable to determine ConfigDir! Is your $HOME set?")
		}

		configDir = filepath.Join(home, ".unisync")

		if !dirExists(configDir) {
			err := os.Mkdir(configDir, 0700)
			if err != nil || !dirExists(configDir) {
				log.Fatalf("Unable to create ConfigDir %v: %v", configDir, err)
			}
		}

	})

	return configDir
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return info.IsDir()
}
