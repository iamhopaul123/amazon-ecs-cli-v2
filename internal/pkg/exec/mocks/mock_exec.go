// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/exec/exec.go

// Package mocks is a generated GoMock package.
package mocks

import (
	command "github.com/aws/copilot-cli/internal/pkg/term/command"
	prompt "github.com/aws/copilot-cli/internal/pkg/term/prompt"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// Mockrunner is a mock of runner interface
type Mockrunner struct {
	ctrl     *gomock.Controller
	recorder *MockrunnerMockRecorder
}

// MockrunnerMockRecorder is the mock recorder for Mockrunner
type MockrunnerMockRecorder struct {
	mock *Mockrunner
}

// NewMockrunner creates a new mock instance
func NewMockrunner(ctrl *gomock.Controller) *Mockrunner {
	mock := &Mockrunner{ctrl: ctrl}
	mock.recorder = &MockrunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Mockrunner) EXPECT() *MockrunnerMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *Mockrunner) Run(name string, args []string, options ...command.Option) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{name, args}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Run", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run
func (mr *MockrunnerMockRecorder) Run(name, args interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{name, args}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*Mockrunner)(nil).Run), varargs...)
}

// Mockprompter is a mock of prompter interface
type Mockprompter struct {
	ctrl     *gomock.Controller
	recorder *MockprompterMockRecorder
}

// MockprompterMockRecorder is the mock recorder for Mockprompter
type MockprompterMockRecorder struct {
	mock *Mockprompter
}

// NewMockprompter creates a new mock instance
func NewMockprompter(ctrl *gomock.Controller) *Mockprompter {
	mock := &Mockprompter{ctrl: ctrl}
	mock.recorder = &MockprompterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Mockprompter) EXPECT() *MockprompterMockRecorder {
	return m.recorder
}

// Confirm mocks base method
func (m *Mockprompter) Confirm(message, help string, promptOpts ...prompt.Option) (bool, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{message, help}
	for _, a := range promptOpts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Confirm", varargs...)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Confirm indicates an expected call of Confirm
func (mr *MockprompterMockRecorder) Confirm(message, help interface{}, promptOpts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{message, help}, promptOpts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Confirm", reflect.TypeOf((*Mockprompter)(nil).Confirm), varargs...)
}