// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/store/store.go

// Package mocks is a generated GoMock package.
package mocks

import (
	identity "github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/identity"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockidentityService is a mock of identityService interface
type MockidentityService struct {
	ctrl     *gomock.Controller
	recorder *MockidentityServiceMockRecorder
}

// MockidentityServiceMockRecorder is the mock recorder for MockidentityService
type MockidentityServiceMockRecorder struct {
	mock *MockidentityService
}

// NewMockidentityService creates a new mock instance
func NewMockidentityService(ctrl *gomock.Controller) *MockidentityService {
	mock := &MockidentityService{ctrl: ctrl}
	mock.recorder = &MockidentityServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockidentityService) EXPECT() *MockidentityServiceMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockidentityService) Get() (identity.Caller, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].(identity.Caller)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockidentityServiceMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockidentityService)(nil).Get))
}
