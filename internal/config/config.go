package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	Cache      CacheConfig `yaml:"cache_config"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost" env:"HTTP_SERVER_ADDRESS"`
	Port        int           `yaml:"port" env-default:"8080" env:"HTTP_SERVER_PORT"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"30s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

type CacheConfig struct {
	Enabled         bool          `yaml:"enabled"`
	TTL             time.Duration `yaml:"ttl" env-default:"30m"`
	ReverseIndexTTL time.Duration `yaml:"reverse_index_ttl" env-default:"30m"`
	PrefixURL       string        `yaml:"prefix_url" env-default:"url:"`
	PrefixRev       string        `yaml:"prefix_rev" env-default:"rev:"`
}

func MustLoad() (*Config, string, string) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	} else {
		log.Println("Using config file:", configPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	db_str := os.Getenv("DATABASE_URL")
	if db_str == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	cache_str := os.Getenv("CACHE_URL")
	if cache_str == "" {
		log.Println("Warning: CACHE_URL is not set")
	}

	return &cfg, db_str, cache_str
}
