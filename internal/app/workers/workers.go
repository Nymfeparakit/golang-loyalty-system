package workers

import (
	"github.com/rs/zerolog/log"
)

const workersNum = 1

func InitOrdersWorkers(
	ordersCh chan string,
	userService UserService,
	orderService OrderService,
	calculator AccrualCalculator,
) {
	log.Info().Msg("starting orders workers")
	for i := 0; i < workersNum; i++ {
		worker := NewOrderAccrualWorker(ordersCh, userService, orderService, calculator)
		go worker.getOrdersAccrual()
	}
}
