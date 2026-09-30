package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ClickhouseConfig struct {
	Addr     string
	Database string
	Username string
	Password string
}

type NatsConfig struct {
	Connstr  string
	Stream   string
	Consumer string
}

type Config struct {
	Nats       NatsConfig
	Clickhouse ClickhouseConfig
}

const (
	defaultNatsConnstr    = "nats://localhost:4222"
	defaultNatsStream     = "wallet"
	defaultNatsConsumer   = "analytics"
	defaultClickhouseAddr = "localhost:9000"
	defaultClickhouseDB   = "default"
	defaultClickhouseUser = "default"
)

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Load(files ...string) (*Config, error) {
	if err := godotenv.Load(files...); err != nil {
		return nil, err
	}
	return &Config{
			Nats: NatsConfig{
				Connstr:  envOrDefault("NATSCONNSTR", defaultNatsConnstr),
				Stream:   envOrDefault("NATSSTREAM", defaultNatsStream),
				Consumer: envOrDefault("NATSCONSUMER", defaultNatsConsumer),
			},
			Clickhouse: ClickhouseConfig{
				Addr:     envOrDefault("CLICKHOUSEADDR", defaultClickhouseAddr),
				Database: envOrDefault("CLICKHOUSEDB", defaultClickhouseDB),
				Username: envOrDefault("CLICKHOUSEUSER", defaultClickhouseUser),
				Password: os.Getenv("CLICKHOUSEPASSWORD"),
			},
		},
		nil
}
