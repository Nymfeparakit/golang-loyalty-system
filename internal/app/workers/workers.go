package workers

import (
	"github.com/jmoiron/sqlx"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/services"
)

func InitWorkers(db *sqlx.DB, config *configs.Config, ordersCh chan string, orderService *services.OrderService) {
	requestsWorker := NewRateLimitedReqWorker()
	go requestsWorker.ProcessRequests()
	accrualCalculator := services.NewAccrualCalculationService(config.AccrualSystemAddr, requestsWorker)

	orderProcessor := NewOrderProcessor()
	orderProcessor.Start(orderService, db, accrualCalculator, ordersCh)
}
