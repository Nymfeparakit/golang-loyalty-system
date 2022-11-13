package workers

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
)

type UserService interface {
	IncreaseBalanceForOrder(ctx context.Context, orderNumber string, accrual int) error
}

type OrderService interface {
	UpdateOrderStatus(ctx context.Context, orderNumber string, orderStatus string) error
}

type AccrualCalculator interface {
	CreateOrderForCalculation(orderNumber string) error
	GetOrderAccrualRes(orderNumber string) (*domain.AccrualCalculationRes, error)
}

type OrderAccrualWorker struct {
	ordersCh          chan string
	userService       UserService
	orderService      OrderService
	accrualCalculator AccrualCalculator
	unprocessedOrders []string
}

func NewOrderAccrualWorker(
	ordersCh chan string,
	userService UserService,
	orderService OrderService,
	accrualCalculator AccrualCalculator,
) *OrderAccrualWorker {
	unprocessedOrders := make([]string, 0)
	return &OrderAccrualWorker{
		ordersCh:          ordersCh,
		accrualCalculator: accrualCalculator,
		userService:       userService,
		orderService:      orderService,
		unprocessedOrders: unprocessedOrders,
	}
}

func (w *OrderAccrualWorker) processOrder(orderNumber string) (bool, error) {
	// получаем сведения по начислению баллов за заказ
	accrualRes, err := w.accrualCalculator.GetOrderAccrualRes(orderNumber)
	if err != nil {
		return false, err
	}

	// обновляем статус заказа
	newOrderStatus := accrualRes.Status
	log.Info().Msg(fmt.Sprintf("updating order status: %v - %v", orderNumber, newOrderStatus))
	err = w.orderService.UpdateOrderStatus(context.Background(), orderNumber, newOrderStatus)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to update order status: %v", err.Error()))
		return false, err
	}

	// todo: remove this line, should take accrual value from system
	//rand.Seed(time.Now().UnixNano())
	//accrualRes.Accrual = rand.Intn(501)
	// если заказ оказался обработанным, то прибавляем пользователю баланс по этому заказу
	if newOrderStatus == domain.OrderProcessedStatus && accrualRes.Accrual != 0 {
		log.Info().Msg(fmt.Sprintf("increasing balance for order '%s', accrual - %d", orderNumber, accrualRes.Accrual))
		err = w.userService.IncreaseBalanceForOrder(context.Background(), orderNumber, accrualRes.Accrual)
		if err != nil {
			log.Error().Msg("increasing user balance failed: " + err.Error())
			return false, err
		}
	}

	if accrualRes.Status == domain.OrderProcessedStatus {
		return true, nil
	}

	return false, nil
}

func (w *OrderAccrualWorker) processOrders() error {
	log.Info().Msg(fmt.Sprintf("processing orders list: %v", w.unprocessedOrders))
	var tmpOrders []string

	for _, orderNumber := range w.unprocessedOrders {
		orderProcessed, err := w.processOrder(orderNumber)
		if err != nil {
			return err
		}
		if !orderProcessed {
			tmpOrders = append(tmpOrders, orderNumber)
		}
	}

	w.unprocessedOrders = tmpOrders

	return nil
}

func (w *OrderAccrualWorker) getOrdersAccrual() {
	for {
		if len(w.unprocessedOrders) == 0 {
			log.Info().Msg("waiting for new order number")
			orderNumber := <-w.ordersCh
			if err := w.accrualCalculator.CreateOrderForCalculation(orderNumber); err != nil {
				log.Error().Msg(fmt.Sprintf("creating order in accrual system failed: %v", err.Error()))
				return
			}
			w.unprocessedOrders = append(w.unprocessedOrders, orderNumber)
		}
		select {
		case orderNumber := <-w.ordersCh:
			log.Info().Msg(fmt.Sprintf("receiving new order '%s' in worker", orderNumber))
			if err := w.accrualCalculator.CreateOrderForCalculation(orderNumber); err != nil {
				log.Error().Msg(fmt.Sprintf("creating order in accrual system failed: %v", err.Error()))
				return
			}
			w.unprocessedOrders = append(w.unprocessedOrders, orderNumber)
		default:
			err := w.processOrders()
			if err != nil {
				log.Error().Msg(fmt.Sprintf("processing orders failed - %v", err.Error()))
				return
			}
		}
	}
}
