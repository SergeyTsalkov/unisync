package client

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"unisync/config"
	"unisync/filelist"
)

type StoredCache struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Host   string `json:"host"`

	List filelist.FileList `json:"list"`
}

func (c *Client) cacheFullpath() string {
	return filepath.Join(config.ConfigDir(), c.Config.Name+".cache")
}

func (c *Client) Cache() (filelist.FileList, error) {
	if c.cache == nil {
		fullpath := c.cacheFullpath()
		bytes, err := os.ReadFile(fullpath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				err = nil
			}
			return nil, err
		}

		stored := &StoredCache{}
		err = json.Unmarshal(bytes, stored)
		if err != nil {
			log.Println("Unable to parse cache file, will generate new one:", fullpath)
			return nil, nil
		}

		if stored.Local != c.GetBasepath() || stored.Remote != c.remoteBasepath || stored.Host != c.Config.Host {
			log.Println("Cache seems to be for a different server, will generate new one:", fullpath)
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
