package internal

import "time"

type PostgresConfig struct {
	Port         int           `yml:"port"`
	Host         string        `yml:"host"`
	User         string        `yml:"user"`
	Retries      int           `yml:"retries"`
	Password     string        `yml:"password"`
	ReadTimeout  time.Duration `yml:"readTimeout"`
	WriteTimeout time.Duration `yml:"writeTimeout"`
	DatabaseName string        `yml:"databaseName"`
	SSLMode      string        `yml:"sslMode"`
}
type Config struct {
	Postgres PostgresConfig `yml:"postgres"`
}
