// Code generated by MockGen. DO NOT EDIT.
// Source: gophermart/internal/app/handlers (interfaces: AuthService)

// Package mock_handlers is a generated GoMock package.
package mock_handlers

import (
	context "context"
	domain "gophermart/internal/app/domain"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockAuthService is a mock of AuthService interface.
type MockAuthService struct {
	ctrl     *gomock.Controller
	recorder *MockAuthServiceMockRecorder
}

// MockAuthServiceMockRecorder is the mock recorder for MockAuthService.
type MockAuthServiceMockRecorder struct {
	mock *MockAuthService
}

// NewMockAuthService creates a new mock instance.
func NewMockAuthService(ctrl *gomock.Controller) *MockAuthService {
	mock := &MockAuthService{ctrl: ctrl}
	mock.recorder = &MockAuthServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthService) EXPECT() *MockAuthServiceMockRecorder {
	return m.recorder
}

// AuthenticateUser mocks base method.
func (m *MockAuthService) AuthenticateUser(arg0 context.Context, arg1, arg2 string) (*domain.TokenData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthenticateUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(*domain.TokenData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthenticateUser indicates an expected call of AuthenticateUser.
func (mr *MockAuthServiceMockRecorder) AuthenticateUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthenticateUser", reflect.TypeOf((*MockAuthService)(nil).AuthenticateUser), arg0, arg1, arg2)
}

// GetUserFromContext mocks base method.
func (m *MockAuthService) GetUserFromContext(arg0 context.Context) (*domain.UserDTO, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserFromContext", arg0)
	ret0, _ := ret[0].(*domain.UserDTO)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetUserFromContext indicates an expected call of GetUserFromContext.
func (mr *MockAuthServiceMockRecorder) GetUserFromContext(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserFromContext", reflect.TypeOf((*MockAuthService)(nil).GetUserFromContext), arg0)
}