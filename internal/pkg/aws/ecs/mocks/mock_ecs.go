// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/aws/ecs/ecs.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	ecs "github.com/aws/copilot-cli/internal/pkg/new-sdk-go/ecs"
	gomock "github.com/golang/mock/gomock"
)

// Mockapi is a mock of api interface.
type Mockapi struct {
	ctrl     *gomock.Controller
	recorder *MockapiMockRecorder
}

// MockapiMockRecorder is the mock recorder for Mockapi.
type MockapiMockRecorder struct {
	mock *Mockapi
}

// NewMockapi creates a new mock instance.
func NewMockapi(ctrl *gomock.Controller) *Mockapi {
	mock := &Mockapi{ctrl: ctrl}
	mock.recorder = &MockapiMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockapi) EXPECT() *MockapiMockRecorder {
	return m.recorder
}

// DescribeClusters mocks base method.
func (m *Mockapi) DescribeClusters(input *ecs.DescribeClustersInput) (*ecs.DescribeClustersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeClusters", input)
	ret0, _ := ret[0].(*ecs.DescribeClustersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeClusters indicates an expected call of DescribeClusters.
func (mr *MockapiMockRecorder) DescribeClusters(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeClusters", reflect.TypeOf((*Mockapi)(nil).DescribeClusters), input)
}

// DescribeServices mocks base method.
func (m *Mockapi) DescribeServices(input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeServices", input)
	ret0, _ := ret[0].(*ecs.DescribeServicesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeServices indicates an expected call of DescribeServices.
func (mr *MockapiMockRecorder) DescribeServices(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeServices", reflect.TypeOf((*Mockapi)(nil).DescribeServices), input)
}

// DescribeTaskDefinition mocks base method.
func (m *Mockapi) DescribeTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeTaskDefinition", input)
	ret0, _ := ret[0].(*ecs.DescribeTaskDefinitionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeTaskDefinition indicates an expected call of DescribeTaskDefinition.
func (mr *MockapiMockRecorder) DescribeTaskDefinition(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeTaskDefinition", reflect.TypeOf((*Mockapi)(nil).DescribeTaskDefinition), input)
}

// DescribeTasks mocks base method.
func (m *Mockapi) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeTasks", input)
	ret0, _ := ret[0].(*ecs.DescribeTasksOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeTasks indicates an expected call of DescribeTasks.
func (mr *MockapiMockRecorder) DescribeTasks(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeTasks", reflect.TypeOf((*Mockapi)(nil).DescribeTasks), input)
}

// ExecuteCommand mocks base method.
func (m *Mockapi) ExecuteCommand(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecuteCommand", input)
	ret0, _ := ret[0].(*ecs.ExecuteCommandOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecuteCommand indicates an expected call of ExecuteCommand.
func (mr *MockapiMockRecorder) ExecuteCommand(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecuteCommand", reflect.TypeOf((*Mockapi)(nil).ExecuteCommand), input)
}

// ListTasks mocks base method.
func (m *Mockapi) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTasks", input)
	ret0, _ := ret[0].(*ecs.ListTasksOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTasks indicates an expected call of ListTasks.
func (mr *MockapiMockRecorder) ListTasks(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTasks", reflect.TypeOf((*Mockapi)(nil).ListTasks), input)
}

// RunTask mocks base method.
func (m *Mockapi) RunTask(input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunTask", input)
	ret0, _ := ret[0].(*ecs.RunTaskOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunTask indicates an expected call of RunTask.
func (mr *MockapiMockRecorder) RunTask(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunTask", reflect.TypeOf((*Mockapi)(nil).RunTask), input)
}

// StopTask mocks base method.
func (m *Mockapi) StopTask(input *ecs.StopTaskInput) (*ecs.StopTaskOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopTask", input)
	ret0, _ := ret[0].(*ecs.StopTaskOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StopTask indicates an expected call of StopTask.
func (mr *MockapiMockRecorder) StopTask(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopTask", reflect.TypeOf((*Mockapi)(nil).StopTask), input)
}

// WaitUntilTasksRunning mocks base method.
func (m *Mockapi) WaitUntilTasksRunning(input *ecs.DescribeTasksInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUntilTasksRunning", input)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUntilTasksRunning indicates an expected call of WaitUntilTasksRunning.
func (mr *MockapiMockRecorder) WaitUntilTasksRunning(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUntilTasksRunning", reflect.TypeOf((*Mockapi)(nil).WaitUntilTasksRunning), input)
}

// MockssmSessionStarter is a mock of ssmSessionStarter interface.
type MockssmSessionStarter struct {
	ctrl     *gomock.Controller
	recorder *MockssmSessionStarterMockRecorder
}

// MockssmSessionStarterMockRecorder is the mock recorder for MockssmSessionStarter.
type MockssmSessionStarterMockRecorder struct {
	mock *MockssmSessionStarter
}

// NewMockssmSessionStarter creates a new mock instance.
func NewMockssmSessionStarter(ctrl *gomock.Controller) *MockssmSessionStarter {
	mock := &MockssmSessionStarter{ctrl: ctrl}
	mock.recorder = &MockssmSessionStarterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockssmSessionStarter) EXPECT() *MockssmSessionStarterMockRecorder {
	return m.recorder
}

// StartSession mocks base method.
func (m *MockssmSessionStarter) StartSession(ssmSession *ecs.Session) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartSession", ssmSession)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartSession indicates an expected call of StartSession.
func (mr *MockssmSessionStarterMockRecorder) StartSession(ssmSession interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartSession", reflect.TypeOf((*MockssmSessionStarter)(nil).StartSession), ssmSession)
}
