// Code generated by MockGen. DO NOT EDIT.
// Source: gophermart/internal/app/handlers (interfaces: RegistrationService)

// Package mock_handlers is a generated GoMock package.
package mock_handlers

import (
	context "context"
	domain "gophermart/internal/app/domain"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRegistrationService is a mock of RegistrationService interface.
type MockRegistrationService struct {
	ctrl     *gomock.Controller
	recorder *MockRegistrationServiceMockRecorder
}

// MockRegistrationServiceMockRecorder is the mock recorder for MockRegistrationService.
type MockRegistrationServiceMockRecorder struct {
	mock *MockRegistrationService
}

// NewMockRegistrationService creates a new mock instance.
func NewMockRegistrationService(ctrl *gomock.Controller) *MockRegistrationService {
	mock := &MockRegistrationService{ctrl: ctrl}
	mock.recorder = &MockRegistrationServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegistrationService) EXPECT() *MockRegistrationServiceMockRecorder {
	return m.recorder
}

// RegisterUser mocks base method.
func (m *MockRegistrationService) RegisterUser(arg0 context.Context, arg1 domain.UserDTO) (*domain.TokenData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterUser", arg0, arg1)
	ret0, _ := ret[0].(*domain.TokenData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterUser indicates an expected call of RegisterUser.
func (mr *MockRegistrationServiceMockRecorder) RegisterUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterUser", reflect.TypeOf((*MockRegistrationService)(nil).RegisterUser), arg0, arg1)
}
