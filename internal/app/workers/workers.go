package workers

import (
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

	log.Info().Msg("starting orders workers")
	for i := 0; i < workersNum; i++ {
		worker := NewOrderAccrualWorker(ordersCh, userService, orderService, accrualCalculator)
		go worker.getOrdersAccrual()
	}

}
