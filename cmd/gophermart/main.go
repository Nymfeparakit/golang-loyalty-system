package main

import (
	"context"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/handlers"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"gophermart/internal/app/workers"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func initOrderService(db *sqlx.DB, orderSender *services.OrderSender) *services.OrderService {
	orderRepository := repositories.NewOrderRepository(db)
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
	orderSender := services.NewOrderSender(ordersCh)
	orderService := initOrderService(db, orderSender)
	// Инициируем хэндлеры для ендпоинтов
	router := handlers.InitRouter(db, cfg, orderService)
	// Запускаем воркеров
	ctx, cancelFunc := context.WithCancel(context.Background())
	runner := workers.NewRunner()
	runner.StartWorkers(ctx, db, cfg, ordersCh, orderService)

	srv := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msg(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)
	<-quitCh
	log.Info().Msg("starting graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info().Msg("waiting for connections to close...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Msg(fmt.Sprintf("server shutdown error: %v", err))
	}

	// после остановки сервера останавливаем отправку ордеров воркерам
	orderSender.Stop()
	// выполняем остановку всех воркеров
	cancelFunc()
	runner.WaitWorkersToStop()
}
