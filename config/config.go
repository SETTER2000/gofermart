package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v7"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type (
	Config struct {
		App     `yaml:"app"`
		HTTP    `yaml:"http"`
		Storage `yaml:"storage"`
		Cookie  `yaml:"cookie"`
		Log     `yaml:"logger"`
	}
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}
	HTTP struct {
		// BASE_URL - базовый адрес результирующего сокращённого URL
		BaseURL string `env:"BASE_URL"`
		// RUN_ADDRESS - адрес запуска HTTP-сервера
		ServerAddress string `env:"RUN_ADDRESS"`
		// адрес системы расчёта начислений: переменная окружения ОС
		Accrual string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	}
	Storage struct {
		// FILE_STORAGE_PATH путь до файла с сокращёнными URL (директории не создаёт)
		FileStorage string `env:"FILE_STORAGE_PATH"`
		// Строка с адресом подключения к БД, например для PostgreSQL (драйвер pgx): postgres://username:password@localhost:5432/database_name
		ConnectDB string `env:"DATABASE_URI"`
	}
	Cookie struct {
		AccessTokenName string `env-required:"true" yaml:"access_token_name" env:"ACCESS_TOKEN_NAME" envDefault:"access_token"`
		SecretKey       string `env-required:"true" yaml:"secret_key" env:"SECRET_KEY" envDefault:"RtsynerpoGIYdab_s234r"` // cookie encryp application
		//ExpirationTime  time.Time `env-required:"true" yaml:"expiration_time" env:"EXPIRATION_TIME"`
	}
	Log struct {
		Level string `env-required:"true" yaml:"log_level"  env:"LOG_LEVEL"`
	}
)

var instance *Config

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	// yaml
	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	// flags
	flag.StringVar(&cfg.HTTP.ServerAddress, "a", "", "host to listen on")
	//flag.StringVar(&cfg.HTTP.ServerAddress, "a", "localhost:8080", "host to listen on")
	flag.StringVar(&cfg.HTTP.Accrual, "r", "", "host to listen on")
	//flag.StringVar(&cfg.HTTP.Accrual, "r", "localhost:8081", "host to listen on")
	flag.StringVar(&cfg.HTTP.BaseURL, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	flag.StringVar(&cfg.Storage.FileStorage, "f", "storage.txt", "path to file with abbreviated URLs")
	flag.StringVar(&cfg.Storage.ConnectDB, "d", "", "dsn connect string urlExample PostgreSQL: postgres://username:password@localhost:5432/database_name")
	flag.StringVar(&cfg.Cookie.SecretKey, "s", "RtsynerpoGIYdab_s234r", "cookie secret key")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Gofermart Version %s %v\nUsage : Project Gofermart - URL Shortener Server\n", os.Args[0], cfg.App.Version)
		flag.PrintDefaults()
	}

	flag.Parse()

	// environ
	err = env.Parse(cfg) // caarlos0
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
