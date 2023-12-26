package config

type Postgres struct {
	Host     string
	Port     int
	User     string
	Password string `json:"-"`
	DbName   string
}
