package services

import (
	"fmt"
	"github.com/rs/zerolog/log"
)

type OrderSender struct {
	ordersCh chan string
}

func NewOrderSender(ordersCh chan string) *OrderSender {
	return &OrderSender{ordersCh: ordersCh}
}

func (s *OrderSender) SendOrderToWorkers(orderNumber string) {
	log.Info().Msg(fmt.Sprintf("sending to workers order '%s'", orderNumber))
	s.ordersCh <- orderNumber
}

func (s *OrderSender) Stop() {
	close(s.ordersCh)
}
