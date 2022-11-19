package main

import (
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"gophermart/internal/app/workers"
)

func initOrderService(db *sqlx.DB, ordersCh chan string) *services.OrderService {
	orderRepository := repositories.NewOrderRepository(db)
	orderSender := services.NewOrderSender(ordersCh)
	return services.NewOrderService(orderRepository, orderSender)
}

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

	ordersCh := make(chan string)
	orderService := initOrderService(db, ordersCh)
	// Инициируем хэндлеры для ендпоинтов
	router := handlers.InitRouter(db, cfg, orderService)
	// Запускаем воркеров
	workers.InitWorkers(db, cfg, ordersCh, orderService)

	router.Run(cfg.RunAddr)
}
