package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"server"`
	Monobank   Monobank `yaml:"monobank"`
}

type HTTPServer struct {
	Host string `yaml:"host"`
	Port string `yaml:"port" env-default:"8080"`
}

type Monobank struct {
	API API `yaml:"api"`
}

type API struct {
	URL string `yaml:"url"`
}

func ReadConfig(configPath string) *Config {
	if configPath == "" {
		log.Fatal("configPath is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
