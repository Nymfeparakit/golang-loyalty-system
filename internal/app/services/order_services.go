package services

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"strconv"
	"time"
)

type OrderRepository interface {
	GetOrCreateOrder(ctx context.Context, orderToCreate domain.OrderDTO) (*domain.OrderDTO, bool, error)
	GetOrdersByUser(ctx context.Context, user *domain.UserDTO) ([]*domain.OrderDTO, error)
	UpdateOrderStatusAndAccrual(ctx context.Context, orderNumber string, orderStatus string, accrual float32) error
	GetOrdersWithStatusesIn(ctx context.Context, statuses []string) ([]*domain.OrderDTO, error)
}

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

type OrderService struct {
	orderRepository OrderRepository
	orderSender     *OrderSender
}

func NewOrderService(orderRepository OrderRepository, orderSender *OrderSender) *OrderService {
	return &OrderService{orderRepository: orderRepository, orderSender: orderSender}
}

func (s *OrderService) GetOrCreateOrder(ctx context.Context, orderToCreate domain.OrderDTO) (*domain.OrderDTO, bool, error) {
	orderToCreate.UploadedAt = time.Now()
	order, created, err := s.orderRepository.GetOrCreateOrder(ctx, orderToCreate)
	if err != nil {
		return nil, false, err
	}
	// проверяем, каким пользователем был создан заказ
	if !created && orderToCreate.UserID != order.UserID {
		return nil, false, ErrOrderExistsForOtherUser
	}

	// отправляем номер для дальнейшей обработки заказа
	if created {
		s.orderSender.SendOrderToWorkers(orderToCreate.Number)
	}

	return order, created, nil
}

func (s *OrderService) GetOrdersByUser(ctx context.Context, user *domain.UserDTO) ([]*domain.OrderDTO, error) {
	return s.orderRepository.GetOrdersByUser(ctx, user)
}

func (s *OrderService) UpdateOrderStatusAndAccrual(ctx context.Context, orderNumber string, orderStatus string, accrual float32) error {
	return s.orderRepository.UpdateOrderStatusAndAccrual(ctx, orderNumber, orderStatus, accrual)
}

func (s *OrderService) GetUnprocessedOrdersNumbers(ctx context.Context) ([]string, error) {
	unprocessedOrders, err := s.orderRepository.GetOrdersWithStatusesIn(
		ctx,
		[]string{domain.OrderNewStatus, domain.OrderProcessingStatus},
	)
	if err != nil {
		return []string{}, err
	}

	var ordersNumbers []string
	for _, order := range unprocessedOrders {
		ordersNumbers = append(ordersNumbers, order.Number)
	}

	return ordersNumbers, nil
}

type OrderNumberValidator struct {
}

func NewOrderNumberValidator() *OrderNumberValidator {
	return &OrderNumberValidator{}
}

func (v *OrderNumberValidator) Validate(orderNumber string) bool {
	isLenEven := (len(orderNumber)-1)%2 == 0
	numbersForSum := make([]int, len(orderNumber))
	for i, r := range orderNumber {
		digit, err := strconv.Atoi(string(r))
		if err != nil {
			return false
		}
		isIndexEven := i%2 == 0
		if isLenEven && !isIndexEven || !isLenEven && isIndexEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		numbersForSum[i] = digit
	}
	sum := 0
	for _, num := range numbersForSum {
		sum += num
	}

	return sum%10 == 0
}
