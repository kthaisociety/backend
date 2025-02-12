package config

type Config struct {
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	Server struct {
		Port string
	}
}

func LoadConfig() (*Config, error) {
	// In a real application, you would load this from environment variables
	// or a configuration file
	cfg := &Config{}
	cfg.Database.Host = "localhost"
	cfg.Database.Port = "5432"
	cfg.Database.User = "postgres"
	cfg.Database.Password = "postgres"
	cfg.Database.DBName = "myapp"
	cfg.Database.SSLMode = "disable"
	cfg.Server.Port = "8080"

	return cfg, nil
}
