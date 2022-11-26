package workers

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"sync"
)

type RegisterOrdersWorker struct {
	registerOrderCh   chan string
	processOrderCh    chan string
	accrualCalculator AccrualCalculator
}

func NewRegisterOrdersWorker(registerOrderCh chan string, processOrderCh chan string, calculator AccrualCalculator) *RegisterOrdersWorker {
	return &RegisterOrdersWorker{registerOrderCh: registerOrderCh, processOrderCh: processOrderCh, accrualCalculator: calculator}
}

func (w *RegisterOrdersWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer close(w.processOrderCh)
	defer wg.Done()
	for {
		select {
		case orderNumber, ok := <-w.registerOrderCh:
			if !ok {
				log.Info().Msg("stopping worker registering orders - register orders channel closed")
				return
			}
			log.Info().Msg(fmt.Sprintf("got new order '%s' for registration", orderNumber))
			err := w.accrualCalculator.CreateOrderForCalculation(orderNumber)
			if err != nil {
				log.Error().Msg(fmt.Sprintf("creating order for accrual failed - %v", err.Error()))
				return
			}
			log.Info().Msg(fmt.Sprintf("sending to accrual workers order '%s'", orderNumber))
			w.processOrderCh <- orderNumber
		case <-ctx.Done():
			log.Info().Msg("stopping worker registering orders - context is done")
			return
		}
	}
}
