package workers

import (
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
)

const accrualWorkersNum = 1
const registerOrderWorkersNum = 1

type OrderProcessor struct {
}

func NewOrderProcessor() *OrderProcessor {
	return &OrderProcessor{}
}

func (p *OrderProcessor) Start(
	orderService *services.OrderService,
	db *sqlx.DB,
	calculator AccrualCalculator,
	ordersCh chan string,
) {
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)

	processOrdersCh := make(chan string)
	log.Info().Msg("starting register orders workers")
	for i := 0; i < registerOrderWorkersNum; i++ {
		worker := NewRegisterOrdersWorker(ordersCh, processOrdersCh, calculator)
		go worker.Run()
	}
	log.Info().Msg("starting orders accrual workers")
	for i := 0; i < accrualWorkersNum; i++ {
		worker := NewOrderAccrualWorker(processOrdersCh, userService, orderService, calculator)
		go worker.Run()
	}
}
