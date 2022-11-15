package workers

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gophermart/internal/app/domain"
	mock_workers "gophermart/internal/app/workers/mocks"
	"testing"
)

func TestOrderAccrualWorker_processOrder(t *testing.T) {
	orderNumber := "123"
	tests := []struct {
		name                  string
		accrualRes            *domain.AccrualCalculationRes
		wantResult            bool
		shouldIncreaseBalance bool
	}{
		{
			name: "order was processed, accrual is 0",
			accrualRes: &domain.AccrualCalculationRes{
				Order:   orderNumber,
				Status:  domain.OrderProcessedStatus,
				Accrual: 0,
			},
			shouldIncreaseBalance: false,
			wantResult:            true,
		},
		{
			name: "order was processed, accrual is > 0",
			accrualRes: &domain.AccrualCalculationRes{
				Order:   orderNumber,
				Status:  domain.OrderProcessedStatus,
				Accrual: 100,
			},
			shouldIncreaseBalance: true,
			wantResult:            true,
		},
		{
			name: "order wasn't processed",
			accrualRes: &domain.AccrualCalculationRes{
				Order:  orderNumber,
				Status: domain.OrderProcessingStatus,
			},
			shouldIncreaseBalance: false,
			wantResult:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаем моки
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			accrualCalculatorMock := mock_workers.NewMockAccrualCalculator(ctrl)
			accrualCalculatorMock.EXPECT().GetOrderAccrualRes(orderNumber).Return(tt.accrualRes, nil)
			userServiceMock := mock_workers.NewMockUserService(ctrl)
			if tt.shouldIncreaseBalance {
				userServiceMock.EXPECT().IncreaseBalanceForOrder(
					gomock.Any(), orderNumber, tt.accrualRes.Accrual,
				).Return(nil)
			}
			orderServiceMock := mock_workers.NewMockOrderService(ctrl)
			orderServiceMock.EXPECT().UpdateOrderStatusAndAccrual(
				gomock.Any(),
				orderNumber,
				tt.accrualRes.Status,
				tt.accrualRes.Accrual,
			).Return(nil)

			orderWorker := NewOrderAccrualWorker(make(chan string), userServiceMock, orderServiceMock, accrualCalculatorMock)
			actualRes, err := orderWorker.processOrder(orderNumber)
			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, actualRes)
		})
	}
}
