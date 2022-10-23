package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unisync/ini"
	"unisync/log"
)

var once sync.Once
var configDir string

type Config struct {
	Name           string   `json:"name"`
	Local          string   `json:"local"`
	Remote         string   `json:"remote"`
	User           string   `json:"user"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	Method         string   `json:"method"`
	Prefer         string   `json:"prefer"'`
	Ignore         []string `json:"ignore"`
	Tmpdir         string   `json:"tmpdir"`
	RemoteTmpdir   string   `json:"remote_tmpdir"`
	SshPath        string   `json:"ssh_path"`
	SshOpts        string   `json:"ssh_opts"`
	SshKeys        []string `json:"ssh_key"`
	Timeout        int      `json:"timeout"`
	ConnectTimeout int      `json:"connect_timeout"`

	RemoteUnisyncPath []string `json:"remote_unisync_path"`

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

func New() *Config {
	config := Config{
		SshPath:        "ssh",
		SshOpts:        "-e none -o BatchMode=yes -o StrictHostKeyChecking=no",
		Prefer:         "newest",
		Timeout:        300,
		ConnectTimeout: 30,
		ChmodLocal:     FileMode{0644},
		ChmodRemote:    FileMode{0644},
		ChmodLocalDir:  FileMode{0755},
		ChmodRemoteDir: FileMode{0755},
		ChmodMask:      FileMode{0100},
		ChmodDirMask:   FileMode{0},
	}

	return &config
}

func Parse(path string) (*Config, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(ConfigDir(), path)
	}

	if !IsFile(path) {
		if !strings.HasSuffix(path, ".conf") {
			path = path + ".conf"
		}

		if !IsFile(path) {
			return nil, fmt.Errorf("ConfigFile %v does not exist", path)
		}
	}
	_, name := filepath.Split(path)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read ConfigFile %v: %v", name, err)
	}

	config := New()
	err = ini.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse ConfigFile %v: %v", name, err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("Problem in ConfigFile %v: %v", name, err)
	}

	config.Name = name
	return config, nil
}

func (c *Config) Validate() error {
	setting_missing_error := "setting %v is required (and missing)"

	if c.Local == "" {
		return fmt.Errorf(setting_missing_error, "local")
	}
	if c.Remote == "" {
		return fmt.Errorf(setting_missing_error, "remote")
	}

	if c.Method == "" {
		if runtime.GOOS == "windows" {
			c.Method = "internalssh"
		} else {
			c.Method = "ssh"
		}
	}
	if err := validateInArray("prefer", c.Prefer, []string{"newest", "oldest", "local", "remote"}); err != nil {
		return err
	}
	if err := validateInArray("method", c.Method, []string{"ssh", "internalssh", "directtls"}); err != nil {
		return err
	}
	if !strings.Contains(c.SshOpts, "-e none") {
		return fmt.Errorf(`setting ssh_opts must contain "-e none"`)
	}
	if c.Method == "internalssh" || c.Method == "ssh" {
		if c.Port == 0 {
			c.Port = 22
		}
		if c.User == "" {
			return fmt.Errorf(setting_missing_error, "user")
		}
		if c.Host == "" {
			return fmt.Errorf(setting_missing_error, "host")
		}
		for _, sshkey := range c.SshKeys {
			if !IsFile(sshkey) {
				return fmt.Errorf("ssh_key=%v <-- file does not exist", sshkey)
			}
		}
	}
	if c.Method == "directtls" && c.Port == 0 {
		return fmt.Errorf(setting_missing_error, "port")
	}
	if c.Method == "internalssh" && len(c.SshKeys) == 0 {
		options := []string{"id_rsa", "id_ecdsa", "id_ed25519", "id_dsa", "identity"}
		for _, option := range options {
			option = filepath.Join(HomeDir(), ".ssh", option)
			if IsFile(option) {
				c.SshKeys = append(c.SshKeys, option)
			}
		}
	}

	if len(c.RemoteUnisyncPath) == 0 {
		c.RemoteUnisyncPath = []string{"unisync", "./unisync"}
	}

	return nil
}

func validateInArray(name, value string, options []string) error {
	for _, option := range options {
		if value == option {
			return nil
		}
	}

	return fmt.Errorf("setting %v must be one of: %v", name, strings.Join(options, ", "))
}

func ConfigDir() string {
	once.Do(func() {
		var err error

		configDir, err = ResolvePath("~/.unisync")
		if err != nil {
			log.Fatalf("ConfigDir error: %v", err)
		}

		err = mkdirIfMissing(configDir, 0700)
		if err != nil {
			log.Fatalf("Unable to create ConfigDir %v: %v", configDir, err)
		}
	})

	return configDir
}

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Your $HOME is not set:", err)
	}
	if home == "" {
		log.Fatalln("Your $HOME is not set")
	}
	return home
}

func ResolvePath(oldpath string) (string, error) {
	newpath := oldpath

	if strings.HasPrefix(newpath, "~/") {
		newpath = strings.Replace(newpath, "~", HomeDir(), 1)
	}

	var err error
	newpath, err = filepath.Abs(newpath)
	if err != nil {
		return "", fmt.Errorf("Unable to resolve %v: %w", oldpath, err)
	}

	return newpath, nil
}

func IsFile(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

func IsDir(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

func mkdirIfMissing(dir string, mode os.FileMode) error {
	if IsDir(dir) {
		return nil
	}
	return os.Mkdir(dir, mode)
}
