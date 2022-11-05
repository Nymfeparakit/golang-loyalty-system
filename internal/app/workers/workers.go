package workers

import (
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/configs"
)

const workersNum = 8

func InitOrdersWorkers(ordersCh chan string, config *configs.Config, userService UserService) {
	log.Info().Msg("starting orders workers")
	for i := 0; i < workersNum; i++ {
		worker := NewOrderAccrualWorker(ordersCh, config.AccrualSystemAddr, userService)
		go worker.getOrdersAccrual()
	}
}
