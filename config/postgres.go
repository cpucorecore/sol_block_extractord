package config

type Postgres struct {
	Host   string
	Port   int
	User   string
	Passwd string `json:"-"`
	Db     string
}
