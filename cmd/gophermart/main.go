package main

import (
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/repositories"
)

func main() {
	// загружаем настройки
	cfg, err := configs.InitConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Подключаемся к БД
	db, err := repositories.InitDB(cfg.DatabaseURI)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	router := handlers.InitRouter(db, cfg)
	router.Run(cfg.RunAddr)
}
