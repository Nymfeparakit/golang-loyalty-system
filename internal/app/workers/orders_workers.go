package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

type UserService interface {
	IncreaseBalanceForOrder(ctx context.Context, orderNumber string, accrual int) error
}

type OrderAccrualWorker struct {
	ordersCh          chan string
	accrualSystemAddr string
	userService       UserService
}

func NewOrderAccrualWorker(ordersCh chan string, accrualSystemAddr string, userService UserService) *OrderAccrualWorker {
	return &OrderAccrualWorker{
		ordersCh:          ordersCh,
		accrualSystemAddr: accrualSystemAddr,
		userService:       userService,
	}
}

const orderProcessedStatus = "PROCESSED"

type AccrualCalculationRes struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

func (w *OrderAccrualWorker) checkOrderAccrual(orderNumber string) (bool, error) {
	log.Info().Msg(fmt.Sprintf("starting processing order '%s'", orderNumber))
	requestURL := w.accrualSystemAddr + "/api/orders/"
	// проверяем, был ли заказ обработан
	res, err := http.Get(requestURL + orderNumber)
	if err != nil {
		log.Error().Msg("request to accrual system failed: " + err.Error())
		return false, err
	}
	if res.StatusCode != http.StatusOK {
		log.Error().Msg("request to accrual system failed: status of response - " + strconv.Itoa(res.StatusCode))
		return false, err
	}

	var accrualRes AccrualCalculationRes
	err = json.NewDecoder(res.Body).Decode(&accrualRes)
	if err != nil {
		log.Error().Msg("request to accrual system failed: " + err.Error())
		return false, err
	}
	// если заказ оказался обработанным, то прибавляем пользователю баланс по этому заказу
	if accrualRes.Status == orderProcessedStatus && accrualRes.Accrual != 0 {
		log.Info().Msg(fmt.Sprintf("increasing balance for order '%s', accrual - %d", orderNumber, accrualRes.Accrual))
		err := w.userService.IncreaseBalanceForOrder(context.Background(), orderNumber, accrualRes.Accrual)
		if err != nil {
			log.Error().Msg("increasing user balance failed: " + err.Error())
			return true, nil
		}
	}

	return false, nil
}

func (w *OrderAccrualWorker) getOrdersAccrual() {
	var unprocessedOrders []string
	for {
		select {
		case orderNumber := <-w.ordersCh:
			log.Info().Msg(fmt.Sprintf("receiving new order '%s' in worker", orderNumber))
			unprocessedOrders = append(unprocessedOrders, orderNumber)
		default:
			var tmpOrders []string
			for _, orderNumber := range unprocessedOrders {
				processed, err := w.checkOrderAccrual(orderNumber)
				if err != nil {
					return
				}
				if !processed {
					tmpOrders = append(tmpOrders, orderNumber)
				}
			}
			unprocessedOrders = tmpOrders
		}
	}
}
