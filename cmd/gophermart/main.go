package main

import (
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/handlers"
)

func initDB(connStr string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", connStr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return db, nil
}

func main() {
	// загружаем настройки
	cfg, err := configs.InitConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Подключаемся к БД
	db, err := initDB(cfg.DatabaseURI)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	router := handlers.InitRouter(db, cfg)
	router.Run(cfg.RunAddr)
}
