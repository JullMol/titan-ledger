package config

import "github.com/spf13/viper"

type Config struct {
	ServerPort string `mapstructure:"SERVER_PORT"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE"`
}

func LoadConfig() (config *Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	
	// Read environment variables directly (for production like Render)
	viper.AutomaticEnv()
	
	// Try to read .env file (for local development), ignore error if not found
	_ = viper.ReadInConfig()

	// Set defaults for production
	viper.SetDefault("SERVER_PORT", "10000")
	viper.SetDefault("DB_SSLMODE", "require")
	
	err = viper.Unmarshal(&config)
	return
}
