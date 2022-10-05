package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	Name   string `json:"name"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Host   string `json:"host"`
	Method string `json:"method"`
	Prefer string `json:"prefer"'`

	Chmod *configTypeChmod `json:"chmod"`
}

type configTypeChmod struct {
	Local     *FileMode `json:"local,omitempty"`
	LocalDir  *FileMode `json:"local_dir,omitempty"`
	Remote    *FileMode `json:"remote,omitempty"`
	RemoteDir *FileMode `json:"remote_dir,omitempty"`
	Mask      *FileMode `json:"mask,omitempty"`
	DirMask   *FileMode `json:"dir_mask,omitempty"`
}

type FileMode struct {
	fs.FileMode
}

func (f *FileMode) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	i, err := strconv.ParseUint(str, 8, 32)
	if err != nil {
		return err
	}
	f.FileMode = fs.FileMode(i)
	return nil
}

func (f *FileMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(f.Perm()), 8))
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

	_, config.Name = filepath.Split(path)
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
		C.Chmod.Local = &FileMode{0644}
	}
	if C.Chmod.LocalDir == nil {
		C.Chmod.LocalDir = &FileMode{0755}
	}
	if C.Chmod.Remote == nil {
		C.Chmod.Remote = &FileMode{0644}
	}
	if C.Chmod.RemoteDir == nil {
		C.Chmod.RemoteDir = &FileMode{0755}
	}
	if C.Chmod.Mask == nil {
		C.Chmod.Mask = &FileMode{0100}
	}
	if C.Chmod.DirMask == nil {
		C.Chmod.DirMask = &FileMode{0}
	}
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
