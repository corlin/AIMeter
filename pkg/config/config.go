package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig   `yaml:"server"`
	Database  DatabaseConfig `yaml:"database"`
	Collector CollectorConfig `yaml:"collector"`
	Rates     RatesConfig    `yaml:"rates"`
}

type ServerConfig struct {
	HTTPPort     int    `yaml:"http_port"`
	GRPCPort     int    `yaml:"grpc_port"`
	OTLPHTTPPort int    `yaml:"otlp_http_port"`
	Env          string `yaml:"env"`
}

type DatabaseConfig struct {
	ClickHouse ClickHouseConfig `yaml:"clickhouse"`
	Postgres   PostgresConfig   `yaml:"postgres"`
}

type ClickHouseConfig struct {
	Addr     string `yaml:"addr"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type PostgresConfig struct {
	DSN string `yaml:"dsn"`
}

type CollectorConfig struct {
	BatchSize       int `yaml:"batch_size"`
	FlushIntervalMs int `yaml:"flush_interval_ms"`
}

type RatesConfig struct {
	SeedFile        string `yaml:"seed_file"`
	CacheTTLMinutes int    `yaml:"cache_ttl_minutes"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			HTTPPort:     8080,
			GRPCPort:     4317,
			OTLPHTTPPort: 4318,
			Env:          "development",
		},
		Database: DatabaseConfig{
			ClickHouse: ClickHouseConfig{
				Addr:     "localhost:9000",
				Database: "aimeter",
				Username: "default",
				Password: "",
			},
			Postgres: PostgresConfig{
				DSN: "postgres://aimeter:aimeter@localhost:5432/aimeter?sslmode=disable",
			},
		},
		Collector: CollectorConfig{
			BatchSize:       500,
			FlushIntervalMs: 200,
		},
		Rates: RatesConfig{
			SeedFile:        "configs/rates_seed.json",
			CacheTTLMinutes: 60,
		},
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Environment variable overrides
	if chAddr := os.Getenv("CLICKHOUSE_ADDR"); chAddr != "" {
		cfg.Database.ClickHouse.Addr = chAddr
	}
	if pgDSN := os.Getenv("POSTGRES_DSN"); pgDSN != "" {
		cfg.Database.Postgres.DSN = pgDSN
	}
	if seedFile := os.Getenv("RATES_SEED_FILE"); seedFile != "" {
		cfg.Rates.SeedFile = seedFile
	}

	return cfg, nil
}
