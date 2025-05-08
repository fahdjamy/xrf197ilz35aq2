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

type LogConfig struct {
	OutputFile string `yaml:"outputFile"`
}

type TimescaleDBConfig struct {
	Port         int           `yml:"port"`
	Host         string        `yml:"host"`
	User         string        `yml:"user"`
	DatabaseName string        `yml:"name"`
	Password     string        `yml:"password"`
	SSLMode      string        `yml:"ssl_mode"`
	ReadTimeout  time.Duration `yml:"read_timeout"`
	WriteTimeout time.Duration `yml:"write_timeout"`
	Retries      int           `yml:"connect_retries"`
}

func (tsDB *TimescaleDBConfig) GetDdURL() (string, error) {
	conn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		tsDB.User,
		tsDB.Password,
		tsDB.Host,
		tsDB.Port,
		tsDB.DatabaseName,
		tsDB.SSLMode,
	)
	return conn, nil
}

type PostgresConfig struct {
	Port         int           `yml:"port"`
	Host         string        `yml:"host"`
	SslMode      bool          `yml:"ssl_mode"`
	User         string        `yml:"user"`
	Name         string        `yml:"db_name"`
	Retries      int           `yml:"retries"`
	Password     string        `yml:"password"`
	ReadTimeout  time.Duration `yml:"read_timeout"`
	WriteTimeout time.Duration `yml:"write_timeout"`
	DatabaseURL  string
}

type RedisConfig struct {
	Address      string        `yaml:"address"`
	Password     string        `yaml:"password"`
	Database     int           `yaml:"database"`
	Protocol     int           `yaml:"protocol"`
	MaxRetries   int           `yaml:"max_retries"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	PoolSize     int           `yaml:"poolSize"`
	MinIdleConns int           `yaml:"minIdleConns"`
}

type Config struct {
	Log         LogConfig         `yml:"log"`
	Redis       RedisConfig       `yml:"redis"`
	Postgres    PostgresConfig    `yml:"postgres"`
	TimescaleDB TimescaleDBConfig `yml:"timescaledb"`
}

var (
	config     *Config
	configOnce sync.Once
	configErr  error
)

func NewConfig(env string) (*Config, error) {
	configOnce.Do(func() {
		yamlFile, err := OpenFromRoot(fmt.Sprintf("configs/config-%s.yml", env))
		if err != nil {
			configErr = fmt.Errorf("failed to open config file: %w :: env=%s", err, env)
			return
		}

		config, err = readConfiguration(yamlFile)
		if err != nil {
			configErr = fmt.Errorf("error reading config file: %w :: env=%s", err, env)
			return
		}
		dbURL, dbURLExists := os.LookupEnv("XRF_Q2_BID_PG_DB_URL")
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

		// IF redis environment variables are set, use environment variables
		redisPortInEnv, portExists := os.LookupEnv("XRF_Q2_REDIS_PORT")
		redisPasswordInEnv, redisPassExists := os.LookupEnv("XRF_Q2_REDIS_PASSWORD")
		redisAddressInEnv, redisAddressExists := os.LookupEnv("XRF_Q2_ADDRESS_PASSWORD")
		if redisAddressExists && portExists && redisPassExists {
			config.Redis.Password = redisPasswordInEnv
			config.Redis.Address = redisAddressInEnv + ":" + redisPortInEnv
		}
	})

	// Important: Check the global error variable *after* once.Do.
	if configErr != nil {
		return nil, configErr // Return the stored error
	}
	return config, nil
}

func createDBURL(pgConfig PostgresConfig) (string, error) {
	if pgConfig.Name == "" {
		return "", fmt.Errorf("dataname cannot be empty")
	}
	conn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%t",
		pgConfig.User,
		pgConfig.Password,
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.Name,
		pgConfig.SslMode,
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

	// NewDecoder returns a new decoder that reads from r (a file)
	decoder := yaml.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
