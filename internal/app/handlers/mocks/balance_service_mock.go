// Code generated by MockGen. DO NOT EDIT.
// Source: gophermart/internal/app/handlers (interfaces: UserBalanceService)

// Package mock_handlers is a generated GoMock package.
package mock_handlers

import (
	context "context"
	domain "gophermart/internal/app/domain"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUserBalanceService is a mock of UserBalanceService interface.
type MockUserBalanceService struct {
	ctrl     *gomock.Controller
	recorder *MockUserBalanceServiceMockRecorder
}

// MockUserBalanceServiceMockRecorder is the mock recorder for MockUserBalanceService.
type MockUserBalanceServiceMockRecorder struct {
	mock *MockUserBalanceService
}

// NewMockUserBalanceService creates a new mock instance.
func NewMockUserBalanceService(ctrl *gomock.Controller) *MockUserBalanceService {
	mock := &MockUserBalanceService{ctrl: ctrl}
	mock.recorder = &MockUserBalanceServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserBalanceService) EXPECT() *MockUserBalanceServiceMockRecorder {
	return m.recorder
}

// GetBalanceAndWithdrawalsSum mocks base method.
func (m *MockUserBalanceService) GetBalanceAndWithdrawalsSum(arg0 context.Context, arg1 int) (*domain.BalanceData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserBalance", arg0, arg1)
	ret0, _ := ret[0].(*domain.BalanceData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalanceAndWithdrawalsSum indicates an expected call of GetBalanceAndWithdrawalsSum.
func (mr *MockUserBalanceServiceMockRecorder) GetBalanceAndWithdrawalsSum(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserBalance", reflect.TypeOf((*MockUserBalanceService)(nil).GetBalanceAndWithdrawalsSum), arg0, arg1)
}

// GetBalanceWithdrawals mocks base method.
func (m *MockUserBalanceService) GetBalanceWithdrawals(arg0 context.Context, arg1 int) ([]*domain.Withdrawal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalanceWithdrawals", arg0, arg1)
	ret0, _ := ret[0].([]*domain.Withdrawal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalanceWithdrawals indicates an expected call of GetBalanceWithdrawals.
func (mr *MockUserBalanceServiceMockRecorder) GetBalanceWithdrawals(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalanceWithdrawals", reflect.TypeOf((*MockUserBalanceService)(nil).GetBalanceWithdrawals), arg0, arg1)
}

// WithdrawBalanceForOrder mocks base method.
func (m *MockUserBalanceService) WithdrawBalanceForOrder(arg0 context.Context, arg1 *domain.OrderDTO, arg2 float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithdrawBalanceForOrder", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// WithdrawBalanceForOrder indicates an expected call of WithdrawBalanceForOrder.
func (mr *MockUserBalanceServiceMockRecorder) WithdrawBalanceForOrder(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithdrawBalanceForOrder", reflect.TypeOf((*MockUserBalanceService)(nil).WithdrawBalanceForOrder), arg0, arg1, arg2)
}
