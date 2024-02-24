package config

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Postgres  Postgres
	Business  Business
	StartSlot uint64
	Workers   uint
}

func (c *Config) ToString() string {
	bytes, _ := json.MarshalIndent(c, "", "  ")
	return string(bytes)
}

var GConfig Config

func LoadFromFile(path string) (err error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(bytes, &GConfig)
	if err != nil {
		return
	}

	return
}
