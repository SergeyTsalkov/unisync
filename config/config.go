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

type configType struct {
	Local  string           `json:"local"`
	Remote string           `json:"remote"`
	Host   string           `json:"host"`
	Method string           `json:"method"`
	Chmod  *configTypeChmod `json:"chmod"`
}

type configTypeChmod struct {
	Local      fileMode `json:"local"`
	LocalDir   fileMode `json:"local_dir"`
	LocalMask  fileMode `json:"local_mask"`
	Remote     fileMode `json:"remote"`
	RemoteDir  fileMode `json:"remote_dir"`
	RemoteMask fileMode `json:"remote_mask"`
}

type fileMode struct {
	fs.FileMode
	isSet bool
}

func (f *fileMode) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	i, err := strconv.ParseUint(str, 8, 32)
	if err != nil {
		return err
	}
	f.FileMode = fs.FileMode(i)
	f.isSet = true
	return nil
}

var C *configType = &configType{}
var once sync.Once
var configDir string

func Parse(path string) error {
	if !filepath.IsAbs(path) {
		path = filepath.Join(ConfigDir(), path)
	}

	if !fileExists(path) {
		if !strings.HasSuffix(path, ".json") {
			path = path + ".json"
		}

		if !fileExists(path) {
			return fmt.Errorf("ConfigFile %v does not exist", path)
		}
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Unable to read ConfigFile %v: %v", path, err)
	}

	err = json.Unmarshal(bytes, C)
	if err != nil || C == nil {
		return fmt.Errorf("Unable to parse ConfigFile %v: %v", path, err)
	}

	applyDefaults()
	return nil
}

func applyDefaults() {
	if C.Chmod == nil {
		C.Chmod = &configTypeChmod{}
	}
	if C.Method == "" {
		C.Method = "ssh"
	}

	if !C.Chmod.Local.isSet {
		C.Chmod.Local.FileMode = 0644
	}
	if !C.Chmod.LocalDir.isSet {
		C.Chmod.LocalDir.FileMode = 0755
	}
	if !C.Chmod.LocalMask.isSet {
		C.Chmod.LocalMask.FileMode = 0100
	}
	if !C.Chmod.Remote.isSet {
		C.Chmod.Remote.FileMode = 0644
	}
	if !C.Chmod.RemoteDir.isSet {
		C.Chmod.RemoteDir.FileMode = 0755
	}
	if !C.Chmod.RemoteMask.isSet {
		C.Chmod.RemoteMask.FileMode = 0100
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
