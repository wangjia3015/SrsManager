package utils

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	datas map[string]interface{}
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) LoadFromFile(fileName string) error {
	if content, err := ioutil.ReadFile(fileName); err != nil {
		return err
	} else if err = json.Unmarshal(content, &c.datas); err != nil {
		return err
	}
	return nil
}

func (c *Config) GetString(key string) string {
	if v, ok := c.datas[key]; ok {
		return v.(string)
	}
	return ""
}

func (c *Config) GetInt(key string) int {
	if v, ok := c.datas[key]; ok {
		return v.(int)
	}
	return -1
}
