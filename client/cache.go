package client

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"unisync/config"
	"unisync/filelist"
	"unisync/log"
)

type StoredCache struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Host   string `json:"host"`

	List filelist.FileList `json:"list"`
}

func (c *Client) cacheFullpath() string {
	name := c.Config.Name
	if name == "" {
		str := fmt.Sprintf("%v:%v:%v", c.GetBasepath(), c.remoteBasepath, c.Config.Host)
		name = fmt.Sprintf("%x", md5.Sum([]byte(str)))
	}

	return filepath.Join(config.ConfigDir(), name+".cache")
}

func (c *Client) Cache() (filelist.FileList, error) {
	if c.cache == nil {
		fullpath := c.cacheFullpath()
		bytes, err := os.ReadFile(fullpath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, nil
			}
			return nil, err
		}

		stored := &StoredCache{}
		err = json.Unmarshal(bytes, stored)
		if err != nil {
			log.Warnln("Unable to parse cache file, will generate new one:", fullpath)
			return nil, nil
		}

		if stored.Local != c.GetBasepath() || stored.Remote != c.remoteBasepath || stored.Host != c.Config.Host {
			log.Warnln("Cache seems to be for a different server, will generate new one:", fullpath)
			return nil, nil
		}

		c.cache = stored.List
	}
	return c.cache, nil
}

func (c *Client) SaveCache(cacheList filelist.FileList) error {
	stored := &StoredCache{
		Local:  c.GetBasepath(),
		Remote: c.remoteBasepath,
		Host:   c.Config.Host,
		List:   cacheList,
	}

	bytes, err := json.Marshal(stored)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.cacheFullpath(), bytes, 0600)
	if err != nil {
		return err
	}

	c.cache = cacheList
	return nil
}

func (c *Client) RemoveCache() {
	c.cache = nil
	os.Remove(c.cacheFullpath())
}
