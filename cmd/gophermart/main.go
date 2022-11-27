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
	"os/signal"
	"syscall"
	"time"
)

func initOrderService(db *sqlx.DB, orderSender services.OrderSender) *services.OrderService {
	orderRepository := repositories.NewOrderRepository(db)
	return services.NewOrderService(orderRepository, orderSender)
}

func initUserService(db *sqlx.DB) *services.UserService {
	userRepository := repositories.NewUserRepository(db)
	return services.NewUserService(userRepository)
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
	accrualService := services.NewAccrualCalculationService(cfg.AccrualSystemAddr, ordersCh)
	orderService := initOrderService(db, accrualService)
	userService := initUserService(db)
	// Инициируем хэндлеры для ендпоинтов
	router := handlers.InitRouter(cfg, orderService, userService)
	// Запускаем воркеров
	ctx, cancelFunc := context.WithCancel(context.Background())
	runner := workers.NewRunner()
	runner.StartWorkers(ctx, ordersCh, orderService, userService, accrualService)

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

	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-notifyCtx.Done()
	stop()
	log.Info().Msg("starting graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info().Msg("waiting for connections to close...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Msg(fmt.Sprintf("server shutdown error: %v", err))
	}

	// после остановки сервера останавливаем отправку ордеров воркерам
	accrualService.Stop()
	// выполняем остановку всех воркеров
	cancelFunc()
	if ok := runner.WaitWorkersToStop(5 * time.Second); !ok {
		log.Error().Msg("waiting for workers to stop - timeout error")
	}
}
