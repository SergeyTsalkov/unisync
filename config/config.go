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
	Name     string   `json:"name"`
	Local    string   `json:"local"`
	Remote   string   `json:"remote"`
	Username string   `json:"username"`
	Host     string   `json:"host"`
	Method   string   `json:"method"`
	Prefer   string   `json:"prefer"'`
	Timeout  int      `json:"timeout"`
	Ignore   []string `json:"ignore"`

	ChmodLocal     FileMode `json:"chmod_local"`
	ChmodLocalDir  FileMode `json:"chmod_local_dir"`
	ChmodRemote    FileMode `json:"chmod_remote"`
	ChmodRemoteDir FileMode `json:"chmod_remote_dir"`
	ChmodMask      FileMode `json:"chmod_mask"`
	ChmodDirMask   FileMode `json:"chmod_dir_mask"`
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

func (f *FileMode) UnmarshalINI(b []byte) error {
	i, err := strconv.ParseUint(string(b), 8, 32)
	if err != nil {
		return err
	}
	f.FileMode = fs.FileMode(i)
	return nil
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

	config := &Config{
		Method:         "ssh",
		Prefer:         "newest",
		ChmodLocal:     FileMode{0644},
		ChmodRemote:    FileMode{0644},
		ChmodLocalDir:  FileMode{0755},
		ChmodRemoteDir: FileMode{0755},
		ChmodMask:      FileMode{0100},
		ChmodDirMask:   FileMode{0},
	}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse ConfigFile %v: %v", path, err)
	}

	_, config.Name = filepath.Split(path)
	err = config.validate()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (C *Config) validate() error {
	if C.Local == "" {
		return fmt.Errorf("config setting local is required (and missing)")
	}
	if C.Remote == "" {
		return fmt.Errorf("config setting remote is required (and missing)")
	}
	if C.Prefer != "newest" && C.Prefer != "oldest" && C.Prefer != "local" && C.Prefer != "remote" {
		return fmt.Errorf("config.prefer must be one of: newest, oldest, local, remote")
	}

	return nil
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
