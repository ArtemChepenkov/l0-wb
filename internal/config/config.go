package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort      string
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	CacheSize       int
	CacheRefreshSec int
	KafkaBrokers    []string
	KafkaTopic      string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{
		ServerPort:      viper.GetString("SERVER_PORT"),
		DBHost:          viper.GetString("DB_HOST"),
		DBPort:          viper.GetInt("DB_PORT"),
		DBUser:          viper.GetString("DB_USER"),
		DBPassword:      viper.GetString("DB_PASSWORD"),
		DBName:          viper.GetString("DB_NAME"),
		CacheSize:       viper.GetInt("CACHE_SIZE"),
		CacheRefreshSec: viper.GetInt("CACHE_REFRESH"),
		KafkaBrokers:    strings.Split(viper.GetString("KAFKA_BROKERS"), ","),
		KafkaTopic:      viper.GetString("KAFKA_TOPIC"),
	}

	if cfg.ServerPort == "" {
		cfg.ServerPort = "8081"
	}
	if cfg.CacheSize == 0 {
		cfg.CacheSize = 100
	}
	if cfg.CacheRefreshSec == 0 {
		cfg.CacheRefreshSec = 30
	}
	if len(cfg.KafkaBrokers) == 0 || cfg.KafkaBrokers[0] == "" {
		cfg.KafkaBrokers = []string{"kafka:9092"}
	}
	if cfg.KafkaTopic == "" {
		cfg.KafkaTopic = "upload-order-topic"
	}
	if cfg.DBHost == "" || cfg.DBName == "" {
		log.Fatalf("db config missing")
	}
	return cfg
}

func (c *Config) PGDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
