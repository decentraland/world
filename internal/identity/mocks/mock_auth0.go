// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/decentraland/world/internal/identity/data (interfaces: IAuth0Service)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	data "github.com/decentraland/world/internal/identity/data"
	gomock "github.com/golang/mock/gomock"
)

// MockIAuth0Service is a mock of IAuth0Service interface
type MockIAuth0Service struct {
	ctrl     *gomock.Controller
	recorder *MockIAuth0ServiceMockRecorder
}

// MockIAuth0ServiceMockRecorder is the mock recorder for MockIAuth0Service
type MockIAuth0ServiceMockRecorder struct {
	mock *MockIAuth0Service
}

// NewMockIAuth0Service creates a new mock instance
func NewMockIAuth0Service(ctrl *gomock.Controller) *MockIAuth0Service {
	mock := &MockIAuth0Service{ctrl: ctrl}
	mock.recorder = &MockIAuth0ServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIAuth0Service) EXPECT() *MockIAuth0ServiceMockRecorder {
	return m.recorder
}

// GetUserInfo mocks base method
func (m *MockIAuth0Service) GetUserInfo(arg0 string) (data.User, error) {
	ret := m.ctrl.Call(m, "GetUserInfo", arg0)
	ret0, _ := ret[0].(data.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserInfo indicates an expected call of GetUserInfo
func (mr *MockIAuth0ServiceMockRecorder) GetUserInfo(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserInfo", reflect.TypeOf((*MockIAuth0Service)(nil).GetUserInfo), arg0)
}