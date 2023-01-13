package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unisync/ini"
	"unisync/log"
)

var once sync.Once
var configDir string

// json is only used to transmit the needed parts of config to server
// ini is used to parse the conf file on the client
type Config struct {
	Name           string        `json:"name" ini:"name"`
	Local          string        `json:"-" ini:"local"`
	Remote         string        `json:"remote" ini:"remote"`
	User           string        `json:"-" ini:"user"`
	Host           string        `json:"-" ini:"host"`
	Port           int           `json:"-" ini:"port"`
	Method         string        `json:"-" ini:"method"`
	Prefer         string        `json:"-"' ini:"prefer"`
	Ignore         []string      `json:"ignore" ini:"ignore"`
	SshPath        string        `json:"-" ini:"ssh_path"`
	SshOpts        string        `json:"-" ini:"ssh_opts"`
	SshKeys        []string      `json:"-" ini:"ssh_key"`
	TlsKey         string        `json:"-" ini:"tls_key"`
	Timeout        time.Duration `json:"-" ini:"timeout"`
	ConnectTimeout time.Duration `json:"-" ini:"connect_timeout"`
	Log            string        `json:"-" ini:"log"`
	Symlinks       bool          `json:"symlinks" ini:"symlinks"`
	Debug          bool          `json:"-" ini:"debug"`

	TmpdirLocal  string `json:"-" ini:"tmpdir_local"`
	TmpdirRemote string `json:"tmpdir_remote" ini:"tmpdir_remote"`

	WatchLocal  string        `json:"-" ini:"watch_local"`
	WatchRemote string        `json:"watch_remote" ini:"watch_remote"`
	PollFreq    time.Duration `json:"poll_freq" ini:"poll_freq"`

	RemoteUnisyncPath []string `json:"-" ini:"remote_unisync_path"`

	ChmodLocal     fs.FileMode `json:"chmod_local" ini:"chmod_local"`
	ChmodLocalDir  fs.FileMode `json:"chmod_local_dir" ini:"chmod_local_dir"`
	ChmodRemote    fs.FileMode `json:"chmod_remote" ini:"chmod_remote"`
	ChmodRemoteDir fs.FileMode `json:"chmod_remote_dir" ini:"chmod_remote_dir"`
	ChmodMask      fs.FileMode `json:"chmod_mask" ini:"chmod_mask"`
	ChmodDirMask   fs.FileMode `json:"chmod_dir_mask" ini:"chmod_dir_mask"`
}

func New(name string) *Config {
	config := Config{
		Name:           name,
		SshPath:        "ssh",
		SshOpts:        "-e none -o BatchMode=yes -o StrictHostKeyChecking=no",
		TlsKey:         "secure.key",
		Prefer:         "newest",
		WatchLocal:     "1",
		WatchRemote:    "1",
		PollFreq:       250 * time.Millisecond,
		Timeout:        300 * time.Second,
		ConnectTimeout: 30 * time.Second,
		ChmodLocal:     0644,
		ChmodRemote:    0644,
		ChmodLocalDir:  0755,
		ChmodRemoteDir: 0755,
		ChmodMask:      0100,
		ChmodDirMask:   0,
	}

	if name != "" {
		config.Log = name + ".log"
	}

	if runtime.GOOS == "windows" {
		config.Method = "internalssh"
		config.Symlinks = false
	} else {
		config.Method = "ssh"
		config.Symlinks = true
	}

	return &config
}

func iniParser() *ini.Parser {
	parseFileMode := func(str string) (reflect.Value, error) {
		i, err := strconv.ParseUint(str, 8, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(fs.FileMode(i)), nil
	}

	parseDuration := func(str string) (reflect.Value, error) {
		i, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(time.Duration(i * float64(time.Second))), nil
	}

	parseIniBool := func(str string) (reflect.Value, error) {
		b, err := parseBool(str)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(b), nil
	}

	parser := ini.New()
	parser.AddTypeMap("fs.FileMode", parseFileMode)
	parser.AddTypeMap("time.Duration", parseDuration)
	parser.AddTypeMap("bool", parseIniBool)
	return parser
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

	config := New(name)
	err = iniParser().Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse ConfigFile %v: %v", name, err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("Problem in ConfigFile %v: %v", name, err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	var err error
	setting_missing_error := "setting %v is required (and missing)"

	if c.Local == "" {
		return fmt.Errorf(setting_missing_error, "local")
	}
	if c.Remote == "" {
		return fmt.Errorf(setting_missing_error, "remote")
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
	if c.WatchLocal, err = validateExtendedBool(c.WatchLocal, "poll"); err != nil {
		return fmt.Errorf("local_watch=%v <-- %v", c.WatchLocal, err)
	}
	if c.WatchRemote, err = validateExtendedBool(c.WatchRemote, "poll"); err != nil {
		return fmt.Errorf("remote_watch=%v <-- %v", c.WatchRemote, err)
	}

	if c.Log != "" && !filepath.IsAbs(c.Log) {
		c.Log = filepath.Join(ConfigDir(), c.Log)
	}

	if len(c.RemoteUnisyncPath) == 0 {
		// .unisync/unisync and ~/.unisync/unisync are not the same
		// initial remote dir on connect is not guaranteed to be $HOME
		c.RemoteUnisyncPath = []string{"unisync", "./unisync", ".unisync/unisync", "~/.unisync/unisync"}
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

func parseBool(str string) (bool, error) {
	str = strings.ToLower(str)
	switch str {
	case "1", "t", "true", "yes", "on":
		return true, nil
	case "0", "f", "false", "no", "off":
		return false, nil
	}

	return false, fmt.Errorf("unable to parse bool")
}

func validateExtendedBool(str string, opts ...string) (string, error) {
	if b, err := parseBool(str); err == nil {
		if b {
			return "1", nil
		}
		return "0", nil
	}

	for _, opt := range opts {
		if str == opt {
			return opt, nil
		}
	}

	return str, fmt.Errorf("unable to parse value")
}
