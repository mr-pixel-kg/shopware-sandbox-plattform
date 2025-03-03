package config

import (
	"github.com/spf13/viper"
	"log"
)

const CONFIG_FILE = "./config.yml"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port           int      `mapstructure:"port"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// LoadConfig loads configuration from a file or from environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigFile(CONFIG_FILE)

	// Init defaults and environment variables
	initConfig()

	// Support environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println(CONFIG_FILE + " not found!")
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func initConfig() {
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.SetDefault("server.port", 8080)
}
