package workers

import (
	"fmt"
	"github.com/rs/zerolog/log"
)

type RegisterOrdersWorker struct {
	registerOrderCh   chan string
	processOrderCh    chan string
	accrualCalculator AccrualCalculator
}

func NewRegisterOrdersWorker(registerOrderCh chan string, processOrderCh chan string, calculator AccrualCalculator) *RegisterOrdersWorker {
	return &RegisterOrdersWorker{registerOrderCh: registerOrderCh, processOrderCh: processOrderCh, accrualCalculator: calculator}
}

func (w *RegisterOrdersWorker) Run() {
	for orderNumber := range w.registerOrderCh {
		log.Info().Msg(fmt.Sprintf("got new order '%s' for registration", orderNumber))
		err := w.accrualCalculator.CreateOrderForCalculation(orderNumber)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("creating order for accrual failed - %v", err.Error()))
			return
		}
		go func(processOrderCh chan string, orderNumber string) {
			log.Info().Msg(fmt.Sprintf("sending to accrual workers order '%s'", orderNumber))
			processOrderCh <- orderNumber
		}(w.processOrderCh, orderNumber)
	}
}
