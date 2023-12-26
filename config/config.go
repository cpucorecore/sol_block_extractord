package config

import "encoding/json"

type Config struct {
	Inj Injective
	Pg  Postgres
	Biz Business
}

func (cfg *Config) ToString() string {
	bs, _ := json.Marshal(cfg)
	return string(bs)
}

var Cfg Config
