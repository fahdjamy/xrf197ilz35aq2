package internal

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"sync"
	"time"
)

const (
	maxRetries = 3
	retryAfter = 2 * time.Second
)

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
	DatabaseURL  string        `yml:"databaseURL"`
}

type Config struct {
	Postgres PostgresConfig `yml:"postgres"`
}

var (
	config     *Config
	configOnce sync.Once
	configErr  error
)

func NewConfig(env string) (*Config, error) {
	configOnce.Do(func() {
		yamlFile, err := os.Open(fmt.Sprintf("configs/%s/config.yml", env))
		if err != nil {
			configErr = fmt.Errorf("failed to open config file: %w :: env=%s", err, env)
			return
		}

		config, err = readConfiguration(yamlFile)
		if err != nil {
			configErr = fmt.Errorf("error reading config file: %w :: env=%s", err, env)
			return
		}
		dbURL, dbURLExists := os.LookupEnv("XRF_BIDDING_PG_DB_URL")
		if dbURLExists {
			config.Postgres.DatabaseURL = dbURL
			return
		}
		dbURL, err = createDBURL(config.Postgres)
		if err != nil {
			configErr = fmt.Errorf("failed to create postgres database URL: %w :: env=%s", err, env)
			return
		}
		config.Postgres.DatabaseURL = dbURL
	})

	// Important: Check the global error variable *after* once.Do.
	if configErr != nil {
		return nil, configErr // Return the stored error
	}
	return config, nil
}

func createDBURL(pgConfig PostgresConfig) (string, error) {
	conn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pgConfig.User,
		pgConfig.Password,
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.DatabaseName,
		pgConfig.SSLMode,
	)
	return conn, nil
}

func readConfiguration(file io.ReadCloser) (*Config, error) {
	defer func() {
		if err := CloseFileWithRetry(file, maxRetries, retryAfter); err != nil {
			fmt.Println(err)
		}
	}()

	// Decode the YAML into a struct
	var config Config

	// NewDecoder returns a new decoder that reads from r (file)
	decoder := yaml.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
