package internal

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	maxRetries    = 3
	retryAfter    = 2 * time.Second
	tsDBEnvURLKey = "XRF_Q2_BID_TS_DB_URL"
	pgDBEnvURLKey = "XRF_Q2_BID_PG_DB_URL"
)

type LogConfig struct {
	OutputFile string `yaml:"outputFile"`
}

type PostgresConfig struct {
	Port         int    `yml:"port"`
	Host         string `yml:"host"`
	User         string `yml:"user"`
	SSLMode      string `yml:"sslMode"`
	Retries      int    `yml:"retries"`
	Password     string `yml:"password"`
	ReadTimeout  int    `yml:"readTimeout"`
	WriteTimeout int    `yml:"writeTimeout"`
	DatabaseName string `yml:"databaseName"`
	MaxPoolConns int    `yml:"maxPoolConns"`
	DatabaseURL  string
}

type RedisConfig struct {
	Address      string `yaml:"address"`
	Password     string `yaml:"password"`
	Database     int    `yaml:"database"`
	Protocol     int    `yaml:"protocol"`
	PoolSize     int    `yaml:"poolSize"`
	MaxRetries   int    `yaml:"maxRetries"`
	DialTimeout  int    `yaml:"dialTimeout"`
	ReadTimeout  int    `yaml:"readTimeout"`
	MinIdleConns int    `yaml:"minIdleConns"`
	WriteTimeout int    `yaml:"writeTimeout"`
}

type Config struct {
	Log         LogConfig      `yml:"log"`
	Redis       RedisConfig    `yml:"redis"`
	Postgres    PostgresConfig `yml:"postgres"`
	TimescaleDB PostgresConfig `yml:"timescaledb"`
}

var (
	config     *Config
	configOnce sync.Once
	configErr  error
)

func loadConfigs(env string) (*Config, error) {
	configOnce.Do(func() {
		configFilePath := fmt.Sprintf("config-%s.yaml", env)
		viper.SetConfigName(configFilePath)
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")

		// AutomaticEnv check for an environment variable any time a viper.Get request is made.

		// Rules: viper checks for an environment variable w/ a name matching the key uppercased and prefixed with the EnvPrefix if set.
		viper.AutomaticEnv()
		viper.SetEnvPrefix("XRF_Q2") // will be uppercased automatically
		// this is useful e.g., want to use . in Get() calls, but environmental variables are to use _ delimiters (e.g., app.port -> APP_PORT)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// Read the config file
		err := viper.ReadInConfig()
		if err != nil {
			configErr = fmt.Errorf("failed to read config file: %w :: env=%s", err, env)
			return
		}

		appConfig := Config{}
		err = viper.Unmarshal(&appConfig)
		if err != nil {
			configErr = fmt.Errorf("failed to unmarshal config file: %w :: env=%s", err, env)
			return
		}

		tsDBURLFromEnv, found := os.LookupEnv(tsDBEnvURLKey)
		if found {
			appConfig.TimescaleDB.DatabaseURL = tsDBURLFromEnv
		} else {
			tsDbURL, err := setTsDbURL(&appConfig.TimescaleDB, env)
			if err != nil {
				configErr = fmt.Errorf("failed to set tsDbURL: %w :: env=%s", err, env)
				return
			}
			appConfig.TimescaleDB.DatabaseURL = tsDbURL
		}
		if !validateDBName(appConfig.TimescaleDB.DatabaseName, "ts") {
			configErr = fmt.Errorf("invalid timescaleDB database name")
		}

		dbURL, dbURLExists := os.LookupEnv(pgDBEnvURLKey)
		if dbURLExists {
			config.Postgres.DatabaseURL = dbURL
		} else {
			pgDbURL, err := setTsDbURL(&appConfig.Postgres, env)
			if err != nil {
				configErr = fmt.Errorf("failed to set pgDbURL: %w :: env=%s", err, env)
				return
			}
			appConfig.Postgres.DatabaseURL = pgDbURL
		}
		if !validateDBName(appConfig.Postgres.DatabaseName, "pg") {
			configErr = fmt.Errorf("invalid postgres database name")
		}

		config = &appConfig
	})

	// Important: Check the global error variable *after* once.Do.
	if configErr != nil {
		return nil, configErr // Return the stored error
	}
	return config, nil
}

func GetConfig(env string) (*Config, error) {
	return loadConfigs(env)
}

func setTsDbURL(tsConfig *PostgresConfig, env string) (string, error) {
	conn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?pool_max_conns=%d",
		tsConfig.User,
		tsConfig.Password,
		tsConfig.Host,
		tsConfig.Port,
		tsConfig.DatabaseName,
		tsConfig.MaxPoolConns,
	)
	if strings.ToLower(env) == strings.ToLower(ProductionEnv) {
		conn = fmt.Sprintf("%s&sslmode=%s", conn, tsConfig.SSLMode)
	}
	return conn, nil
}

func validateDBName(name string, expectedSubStr string) bool {
	return name != "" && strings.Contains(name, expectedSubStr)
}
