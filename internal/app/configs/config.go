package configs

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	AccrualSystemAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	RunAddr           string `env:"RUN_ADDRESS"`
	DatabaseURI       string `env:"DATABASE_URI"`
}

// InitFlags иницирует флаги, используемые при запуске сервера
func InitFlags(cfg *Config) {
	flag.StringVar(
		&cfg.AccrualSystemAddr, "r", cfg.AccrualSystemAddr, "Address of the accrual calculation system",
	)
	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "Service address and port")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "Database connection address")
}

func InitConfig() (*Config, error) {
	// Загружаем переменные окружения
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	// Инициируем флаги
	InitFlags(&cfg)
	// Переписываем содержимое конфигна значениями из переданных флагов
	flag.Parse()
	return &cfg, nil
}
