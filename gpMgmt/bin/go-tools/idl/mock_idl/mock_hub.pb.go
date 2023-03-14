// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/greenplum-db/gpdb/gp/idl (interfaces: HubClient,HubServer)

// Package mock_idl is a generated GoMock package.
package mock_idl

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	idl "github.com/greenplum-db/gpdb/gp/idl"
	grpc "google.golang.org/grpc"
)

// MockHubClient is a mock of HubClient interface.
type MockHubClient struct {
	ctrl     *gomock.Controller
	recorder *MockHubClientMockRecorder
}

// MockHubClientMockRecorder is the mock recorder for MockHubClient.
type MockHubClientMockRecorder struct {
	mock *MockHubClient
}

// NewMockHubClient creates a new mock instance.
func NewMockHubClient(ctrl *gomock.Controller) *MockHubClient {
	mock := &MockHubClient{ctrl: ctrl}
	mock.recorder = &MockHubClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHubClient) EXPECT() *MockHubClientMockRecorder {
	return m.recorder
}

// StartAgents mocks base method.
func (m *MockHubClient) StartAgents(arg0 context.Context, arg1 *idl.StartAgentsRequest, arg2 ...grpc.CallOption) (*idl.StartAgentsReply, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StartAgents", varargs...)
	ret0, _ := ret[0].(*idl.StartAgentsReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartAgents indicates an expected call of StartAgents.
func (mr *MockHubClientMockRecorder) StartAgents(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartAgents", reflect.TypeOf((*MockHubClient)(nil).StartAgents), varargs...)
}

// StatusAgents mocks base method.
func (m *MockHubClient) StatusAgents(arg0 context.Context, arg1 *idl.StatusAgentsRequest, arg2 ...grpc.CallOption) (*idl.StatusAgentsReply, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StatusAgents", varargs...)
	ret0, _ := ret[0].(*idl.StatusAgentsReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StatusAgents indicates an expected call of StatusAgents.
func (mr *MockHubClientMockRecorder) StatusAgents(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatusAgents", reflect.TypeOf((*MockHubClient)(nil).StatusAgents), varargs...)
}

// Stop mocks base method.
func (m *MockHubClient) Stop(arg0 context.Context, arg1 *idl.StopHubRequest, arg2 ...grpc.CallOption) (*idl.StopHubReply, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Stop", varargs...)
	ret0, _ := ret[0].(*idl.StopHubReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stop indicates an expected call of Stop.
func (mr *MockHubClientMockRecorder) Stop(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockHubClient)(nil).Stop), varargs...)
}

// StopAgents mocks base method.
func (m *MockHubClient) StopAgents(arg0 context.Context, arg1 *idl.StopAgentsRequest, arg2 ...grpc.CallOption) (*idl.StopAgentsReply, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StopAgents", varargs...)
	ret0, _ := ret[0].(*idl.StopAgentsReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StopAgents indicates an expected call of StopAgents.
func (mr *MockHubClientMockRecorder) StopAgents(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopAgents", reflect.TypeOf((*MockHubClient)(nil).StopAgents), varargs...)
}

// MockHubServer is a mock of HubServer interface.
type MockHubServer struct {
	ctrl     *gomock.Controller
	recorder *MockHubServerMockRecorder
}

// MockHubServerMockRecorder is the mock recorder for MockHubServer.
type MockHubServerMockRecorder struct {
	mock *MockHubServer
}

// NewMockHubServer creates a new mock instance.
func NewMockHubServer(ctrl *gomock.Controller) *MockHubServer {
	mock := &MockHubServer{ctrl: ctrl}
	mock.recorder = &MockHubServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHubServer) EXPECT() *MockHubServerMockRecorder {
	return m.recorder
}

// StartAgents mocks base method.
func (m *MockHubServer) StartAgents(arg0 context.Context, arg1 *idl.StartAgentsRequest) (*idl.StartAgentsReply, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartAgents", arg0, arg1)
	ret0, _ := ret[0].(*idl.StartAgentsReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartAgents indicates an expected call of StartAgents.
func (mr *MockHubServerMockRecorder) StartAgents(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartAgents", reflect.TypeOf((*MockHubServer)(nil).StartAgents), arg0, arg1)
}

// StatusAgents mocks base method.
func (m *MockHubServer) StatusAgents(arg0 context.Context, arg1 *idl.StatusAgentsRequest) (*idl.StatusAgentsReply, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StatusAgents", arg0, arg1)
	ret0, _ := ret[0].(*idl.StatusAgentsReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StatusAgents indicates an expected call of StatusAgents.
func (mr *MockHubServerMockRecorder) StatusAgents(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatusAgents", reflect.TypeOf((*MockHubServer)(nil).StatusAgents), arg0, arg1)
}

// Stop mocks base method.
func (m *MockHubServer) Stop(arg0 context.Context, arg1 *idl.StopHubRequest) (*idl.StopHubReply, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", arg0, arg1)
	ret0, _ := ret[0].(*idl.StopHubReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stop indicates an expected call of Stop.
func (mr *MockHubServerMockRecorder) Stop(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockHubServer)(nil).Stop), arg0, arg1)
}

// StopAgents mocks base method.
func (m *MockHubServer) StopAgents(arg0 context.Context, arg1 *idl.StopAgentsRequest) (*idl.StopAgentsReply, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopAgents", arg0, arg1)
	ret0, _ := ret[0].(*idl.StopAgentsReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StopAgents indicates an expected call of StopAgents.
func (mr *MockHubServerMockRecorder) StopAgents(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopAgents", reflect.TypeOf((*MockHubServer)(nil).StopAgents), arg0, arg1)
}
