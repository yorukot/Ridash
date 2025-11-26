package config

import (
	"sync"

	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
)

type AppEnv string

const (
	AppEnvDev  AppEnv = "dev"
	AppEnvProd AppEnv = "prod"
)

// EnvConfig holds all environment variables for the application
type EnvConfig struct {
	// PostgreSQL Settings
	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBSSLMode  string `env:"DB_SSL_MODE,required"`

	AppEnv       AppEnv `env:"APP_ENV" envDefault:"prod"`
	AppName      string `env:"APP_NAME" envDefault:"knocker"`
	AppMachineID int16  `env:"APP_MACHINE_ID" envDefault:"1"`
	AppPort      string `env:"APP_PORT" envDefault:"8080"`

	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	GoogleRedirectURL  string `env:"GOOGLE_REDIRECT_URL,required"`

	// Optional Settings
	OAuthStateExpiresAt   int `env:"OAUTH_STATE_EXPIRES_AT" envDefault:"600"`        // 10 minutes
	AccessTokenExpiresAt  int `env:"ACCESS_TOKEN_EXPIRES_AT" envDefault:"900"`       // 15 minutes
	RefreshTokenExpiresAt int `env:"REFRESH_TOKEN_EXPIRES_AT" envDefault:"31536000"` // 365 days

	JWTSecretKey   string `env:"JWT_SECRET_KEY,required" envDefault:"change_me_to_a_secure_key"`
	FrontendDomain string `env:"FRONTEND_DOMAIN" envDefault:"http://localhost:3000"`
}

var (
	appConfig *EnvConfig
	once      sync.Once
)

// loadConfig loads and validates all environment variables
func loadConfig() (*EnvConfig, error) {
	cfg := &EnvConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// InitConfig initializes the config only once
func InitConfig() (*EnvConfig, error) {
	var err error
	once.Do(func() {
		appConfig, err = loadConfig()
	})
	return appConfig, err
}

// Env returns the config. Panics if not initialized.
func Env() *EnvConfig {
	if appConfig == nil {
		panic("config not initialized — call InitConfig() first")
	}
	return appConfig
}
