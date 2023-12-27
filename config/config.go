package config

import "encoding/json"

type Config struct {
	Pg           Postgres
	Biz          Business
	StartSlot    uint64
	BlockWorkers int
}

func (cfg *Config) ToString() string {
	bs, _ := json.Marshal(cfg)
	return string(bs)
}

var Cfg Config
