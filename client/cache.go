package client

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"unisync/config"
	"unisync/filelist"
)

func (c *Client) Cache() (filelist.FileList, error) {
	if c.cache == nil {
		bytes, err := os.ReadFile(c.cacheFilename())
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				err = nil
			}
			return nil, err
		}

		err = json.Unmarshal(bytes, c.cache)
		if err != nil {
			return nil, err
		}
	}
	return c.cache, nil
}

func (c *Client) cacheFilename() string {
	return filepath.Join(config.ConfigDir(), c.Config.Name+".cache")
}

func (c *Client) SaveCache(cacheList filelist.FileList) error {
	bytes, err := json.Marshal(cacheList)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.cacheFilename(), bytes, 0600)
	if err != nil {
		return err
	}

	c.cache = cacheList
	return nil
}
