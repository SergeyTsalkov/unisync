package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unisync/ini"
	"unisync/log"
)

var once sync.Once
var configDir string

type Config struct {
	Name         string   `json:"name"`
	Local        string   `json:"local"`
	Remote       string   `json:"remote"`
	Username     string   `json:"username"`
	Host         string   `json:"host"`
	Method       string   `json:"method"`
	Prefer       string   `json:"prefer"'`
	Ignore       []string `json:"ignore"`
	Tmpdir       string   `json:"tmpdir"`
	RemoteTmpdir string   `json:"remote_tmpdir"`
	SshPath      string   `json:"ssh_path"`
	SshOpts      string   `json:"ssh_opts"`
	SshKey       string   `json:"ssh_key"`

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
		SshOpts:        "-e none -o BatchMode=yes -o ConnectTimeout=30 -o StrictHostKeyChecking=no",
		Method:         "ssh",
		Prefer:         "newest",
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
	if c.Local == "" {
		return fmt.Errorf("setting local is required (and missing)")
	}
	if c.Remote == "" {
		return fmt.Errorf("setting remote is required (and missing)")
	}
	if err := validateInArray("prefer", c.Prefer, []string{"newest", "oldest", "local", "remote"}); err != nil {
		return err
	}
	if err := validateInArray("method", c.Method, []string{"ssh", "internalssh"}); err != nil {
		return err
	}
	if !strings.Contains(c.SshOpts, "-e none") {
		return fmt.Errorf(`setting ssh_opts must contain "-e none"`)
	}
	if c.Method == "internalssh" && c.SshKey == "" {
		return fmt.Errorf("if you use method=internalssh, ssh_key= must also be set")
	}

	if len(c.RemoteUnisyncPath) == 0 {
		c.RemoteUnisyncPath = []string{"unisync", "./unisync"}
	}

	return nil
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

func ResolvePath(oldpath string) (string, error) {
	newpath := oldpath

	if strings.HasPrefix(newpath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("Unable to resolve %v: %w", oldpath, err)
		}
		if home == "" {
			return "", fmt.Errorf("Unable to resolve %v: your $HOME is not set", oldpath)
		}

		newpath = strings.Replace(newpath, "~", home, 1)
	}

	var err error
	newpath, err = filepath.Abs(newpath)
	if err != nil {
		return "", fmt.Errorf("Unable to resolve %v: %w", oldpath, err)
	}

	return newpath, nil
}

func validateInArray(name, value string, options []string) error {
	for _, option := range options {
		if value == option {
			return nil
		}
	}

	return fmt.Errorf("setting %v must be one of: %v", name, strings.Join(options, ", "))
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
