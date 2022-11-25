package workers

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/domain"
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
}

func NewRunner() *Runner {
	return &Runner{ordersWorkersWG: &sync.WaitGroup{}}
}

func (r *Runner) StartWorkers(
	ctx context.Context,
	config *configs.Config,
	ordersCh chan string,
	orderService *services.OrderService,
	userService *services.UserService,
) {
	accrualCalculator := services.NewAccrualCalculationService(config.AccrualSystemAddr)

	processOrdersCh := make(chan string, 100)
	log.Info().Msg("starting register orders worker")
	worker := NewRegisterOrdersWorker(ordersCh, processOrdersCh, accrualCalculator)
	r.ordersWorkersWG.Add(1)
	go worker.Run(ctx, r.ordersWorkersWG)

	log.Info().Msg("starting orders accrual workers")
	orders, err := orderService.GetUnprocessedOrdersNumbers(ctx)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("getting unprocessed orders failed - %v", err.Error()))
		return
	}

	ordersNum := len(orders)
	ordersPerWorker := ordersNum / accrualWorkersNum
	for i := 0; i < accrualWorkersNum; i++ {
		var ordersForWorker []string
		ordersStartIdx := i * ordersPerWorker
		if i == accrualWorkersNum-1 {
			ordersForWorker = orders[ordersStartIdx:]
		} else {
			ordersForWorker = orders[ordersStartIdx : ordersStartIdx+ordersPerWorker]
		}
		worker := NewOrderAccrualWorker(processOrdersCh, userService, orderService, accrualCalculator, ordersForWorker)
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
	}()
	select {
	case <-notifyCh:
		return true
	case <-time.After(timeout):
		return false
	}
}
