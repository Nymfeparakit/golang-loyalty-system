package workers

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
)

const workersNum = 1

func InitWorkers(db *sqlx.DB, config *configs.Config, ordersCh chan string, orderService *services.OrderService) {
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)

	requestsWorker := NewRateLimitedReqWorker()
	go requestsWorker.ProcessRequests()
	accrualCalculator := services.NewAccrualCalculationService(config.AccrualSystemAddr, requestsWorker)

	orders, err := orderService.GetUnprocessedOrdersNumbers(context.Background())
	if err != nil {
		log.Error().Msg(fmt.Sprintf("getting unprocessed orders failed - %v", err.Error()))
		return
	}
	ordersNum := len(orders)
	ordersPerWorker := ordersNum / workersNum
	log.Info().Msg("starting orders workers")
	for i := 0; i < workersNum; i++ {
		var ordersForWorker []string
		ordersStartIdx := i * ordersPerWorker
		if i == workersNum-1 {
			ordersForWorker = orders[ordersStartIdx:]
		} else {
			ordersForWorker = orders[ordersStartIdx : ordersStartIdx+ordersPerWorker]
		}
		worker := NewOrderAccrualWorker(ordersCh, userService, orderService, accrualCalculator, ordersForWorker)
		go worker.getOrdersAccrual()
	}
}
