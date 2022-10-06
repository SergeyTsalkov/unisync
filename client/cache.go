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

type Cache struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Host   string `json:"host"`

	List filelist.FileList `json:"list"`
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

		err = json.Unmarshal(bytes, &c.cache)
		if err != nil {
			log.Println("Unable to parse cache file, proceeding without:", fullpath)
			return nil, nil
		}
	}
	return c.cache, nil
}

func (c *Client) cacheFullpath() string {
	return filepath.Join(config.ConfigDir(), c.Config.Name+".cache")
}

func (c *Client) SaveCache(cacheList filelist.FileList) error {
	bytes, err := json.Marshal(cacheList)
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
