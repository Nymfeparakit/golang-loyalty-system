package workers

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"sync"
	"time"
)

const (
	accrualWorkersNum = 2
)

type AccrualCalculator interface {
	CreateOrderForCalculation(ctx context.Context, orderNumber string) error
	GetOrderAccrualRes(ctx context.Context, orderNumber string) (*domain.AccrualCalculationRes, error)
}

type Runner struct {
	ordersWorkersWG *sync.WaitGroup
	requestWorkerWG *sync.WaitGroup
}

func NewRunner() *Runner {
	return &Runner{ordersWorkersWG: &sync.WaitGroup{}, requestWorkerWG: &sync.WaitGroup{}}
}

func (r *Runner) StartWorkers(
	ctx context.Context, db *sqlx.DB, config *configs.Config, ordersCh chan string, orderService *services.OrderService,
) {
	requestsWorker := NewRateLimitedReqWorker()
	r.requestWorkerWG.Add(1)
	go requestsWorker.Run(ctx, r.requestWorkerWG)
	accrualCalculator := services.NewAccrualCalculationService(config.AccrualSystemAddr, requestsWorker)

	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)

	processOrdersCh := make(chan string)
	log.Info().Msg("starting register orders worker")
	worker := NewRegisterOrdersWorker(ordersCh, processOrdersCh, accrualCalculator)
	r.ordersWorkersWG.Add(1)
	go worker.Run(ctx, r.ordersWorkersWG)

	log.Info().Msg("starting orders accrual workers")
	for i := 0; i < accrualWorkersNum; i++ {
		worker := NewOrderAccrualWorker(processOrdersCh, userService, orderService, accrualCalculator)
		r.ordersWorkersWG.Add(1)
		go worker.Run(ctx, r.ordersWorkersWG)
	}
}

func (r *Runner) WaitWorkersToStop(timeout time.Duration) bool {
	notifyCh := make(chan struct{})
	go func() {
		defer close(notifyCh)
		log.Info().Msg("waiting orders workers to stop...")
		r.ordersWorkersWG.Wait()
		log.Info().Msg("all orders workers stopped!")
		log.Info().Msg("waiting for request worker to stop...")
		r.requestWorkerWG.Wait()
		log.Info().Msg("request worker stopped!")
	}()
	select {
	case <-notifyCh:
		return true
	case <-time.After(timeout):
		return false
	}
}
