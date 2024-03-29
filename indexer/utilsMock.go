// Code generated by MockGen. DO NOT EDIT.
// Source: utils.go

// Package indexer is a generated GoMock package.
package indexer

import (
	url "net/url"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIURLResolver is a mock of IURLResolver interface.
type MockIURLResolver struct {
	ctrl     *gomock.Controller
	recorder *MockIURLResolverMockRecorder
}

// MockIURLResolverMockRecorder is the mock recorder for MockIURLResolver.
type MockIURLResolverMockRecorder struct {
	mock *MockIURLResolver
}

// NewMockIURLResolver creates a new mock instance.
func NewMockIURLResolver(ctrl *gomock.Controller) *MockIURLResolver {
	mock := &MockIURLResolver{ctrl: ctrl}
	mock.recorder = &MockIURLResolverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIURLResolver) EXPECT() *MockIURLResolverMockRecorder {
	return m.recorder
}

// Resolve mocks base method.
func (m *MockIURLResolver) Resolve(partialURL string) (*url.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Resolve", partialURL)
	ret0, _ := ret[0].(*url.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Resolve indicates an expected call of Resolve.
func (mr *MockIURLResolverMockRecorder) Resolve(partialURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resolve", reflect.TypeOf((*MockIURLResolver)(nil).Resolve), partialURL)
}
