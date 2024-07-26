// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	datagateway "github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	entity "github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"

	mock "github.com/stretchr/testify/mock"
)

// NodeSaleDataGatewayWithTx is an autogenerated mock type for the NodeSaleDataGatewayWithTx type
type NodeSaleDataGatewayWithTx struct {
	mock.Mock
}

type NodeSaleDataGatewayWithTx_Expecter struct {
	mock *mock.Mock
}

func (_m *NodeSaleDataGatewayWithTx) EXPECT() *NodeSaleDataGatewayWithTx_Expecter {
	return &NodeSaleDataGatewayWithTx_Expecter{mock: &_m.Mock}
}

// BeginNodeSaleTx provides a mock function with given fields: ctx
func (_m *NodeSaleDataGatewayWithTx) BeginNodeSaleTx(ctx context.Context) (datagateway.NodeSaleDataGatewayWithTx, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for BeginNodeSaleTx")
	}

	var r0 datagateway.NodeSaleDataGatewayWithTx
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (datagateway.NodeSaleDataGatewayWithTx, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) datagateway.NodeSaleDataGatewayWithTx); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datagateway.NodeSaleDataGatewayWithTx)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BeginNodeSaleTx'
type NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call struct {
	*mock.Call
}

// BeginNodeSaleTx is a helper method to define mock.On call
//   - ctx context.Context
func (_e *NodeSaleDataGatewayWithTx_Expecter) BeginNodeSaleTx(ctx interface{}) *NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call {
	return &NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call{Call: _e.mock.On("BeginNodeSaleTx", ctx)}
}

func (_c *NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call) Run(run func(ctx context.Context)) *NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call) Return(_a0 datagateway.NodeSaleDataGatewayWithTx, _a1 error) *NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call) RunAndReturn(run func(context.Context) (datagateway.NodeSaleDataGatewayWithTx, error)) *NodeSaleDataGatewayWithTx_BeginNodeSaleTx_Call {
	_c.Call.Return(run)
	return _c
}

// ClearDelegate provides a mock function with given fields: ctx
func (_m *NodeSaleDataGatewayWithTx) ClearDelegate(ctx context.Context) (int64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ClearDelegate")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (int64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_ClearDelegate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClearDelegate'
type NodeSaleDataGatewayWithTx_ClearDelegate_Call struct {
	*mock.Call
}

// ClearDelegate is a helper method to define mock.On call
//   - ctx context.Context
func (_e *NodeSaleDataGatewayWithTx_Expecter) ClearDelegate(ctx interface{}) *NodeSaleDataGatewayWithTx_ClearDelegate_Call {
	return &NodeSaleDataGatewayWithTx_ClearDelegate_Call{Call: _e.mock.On("ClearDelegate", ctx)}
}

func (_c *NodeSaleDataGatewayWithTx_ClearDelegate_Call) Run(run func(ctx context.Context)) *NodeSaleDataGatewayWithTx_ClearDelegate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_ClearDelegate_Call) Return(_a0 int64, _a1 error) *NodeSaleDataGatewayWithTx_ClearDelegate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_ClearDelegate_Call) RunAndReturn(run func(context.Context) (int64, error)) *NodeSaleDataGatewayWithTx_ClearDelegate_Call {
	_c.Call.Return(run)
	return _c
}

// Commit provides a mock function with given fields: ctx
func (_m *NodeSaleDataGatewayWithTx) Commit(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeSaleDataGatewayWithTx_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type NodeSaleDataGatewayWithTx_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
//   - ctx context.Context
func (_e *NodeSaleDataGatewayWithTx_Expecter) Commit(ctx interface{}) *NodeSaleDataGatewayWithTx_Commit_Call {
	return &NodeSaleDataGatewayWithTx_Commit_Call{Call: _e.mock.On("Commit", ctx)}
}

func (_c *NodeSaleDataGatewayWithTx_Commit_Call) Run(run func(ctx context.Context)) *NodeSaleDataGatewayWithTx_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_Commit_Call) Return(_a0 error) *NodeSaleDataGatewayWithTx_Commit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_Commit_Call) RunAndReturn(run func(context.Context) error) *NodeSaleDataGatewayWithTx_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// CreateBlock provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) CreateBlock(ctx context.Context, arg entity.Block) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for CreateBlock")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entity.Block) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeSaleDataGatewayWithTx_CreateBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateBlock'
type NodeSaleDataGatewayWithTx_CreateBlock_Call struct {
	*mock.Call
}

// CreateBlock is a helper method to define mock.On call
//   - ctx context.Context
//   - arg entity.Block
func (_e *NodeSaleDataGatewayWithTx_Expecter) CreateBlock(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_CreateBlock_Call {
	return &NodeSaleDataGatewayWithTx_CreateBlock_Call{Call: _e.mock.On("CreateBlock", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_CreateBlock_Call) Run(run func(ctx context.Context, arg entity.Block)) *NodeSaleDataGatewayWithTx_CreateBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(entity.Block))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateBlock_Call) Return(_a0 error) *NodeSaleDataGatewayWithTx_CreateBlock_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateBlock_Call) RunAndReturn(run func(context.Context, entity.Block) error) *NodeSaleDataGatewayWithTx_CreateBlock_Call {
	_c.Call.Return(run)
	return _c
}

// CreateEvent provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) CreateEvent(ctx context.Context, arg entity.NodeSaleEvent) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for CreateEvent")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entity.NodeSaleEvent) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeSaleDataGatewayWithTx_CreateEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateEvent'
type NodeSaleDataGatewayWithTx_CreateEvent_Call struct {
	*mock.Call
}

// CreateEvent is a helper method to define mock.On call
//   - ctx context.Context
//   - arg entity.NodeSaleEvent
func (_e *NodeSaleDataGatewayWithTx_Expecter) CreateEvent(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_CreateEvent_Call {
	return &NodeSaleDataGatewayWithTx_CreateEvent_Call{Call: _e.mock.On("CreateEvent", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_CreateEvent_Call) Run(run func(ctx context.Context, arg entity.NodeSaleEvent)) *NodeSaleDataGatewayWithTx_CreateEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(entity.NodeSaleEvent))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateEvent_Call) Return(_a0 error) *NodeSaleDataGatewayWithTx_CreateEvent_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateEvent_Call) RunAndReturn(run func(context.Context, entity.NodeSaleEvent) error) *NodeSaleDataGatewayWithTx_CreateEvent_Call {
	_c.Call.Return(run)
	return _c
}

// CreateNode provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) CreateNode(ctx context.Context, arg entity.Node) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for CreateNode")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entity.Node) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeSaleDataGatewayWithTx_CreateNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateNode'
type NodeSaleDataGatewayWithTx_CreateNode_Call struct {
	*mock.Call
}

// CreateNode is a helper method to define mock.On call
//   - ctx context.Context
//   - arg entity.Node
func (_e *NodeSaleDataGatewayWithTx_Expecter) CreateNode(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_CreateNode_Call {
	return &NodeSaleDataGatewayWithTx_CreateNode_Call{Call: _e.mock.On("CreateNode", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_CreateNode_Call) Run(run func(ctx context.Context, arg entity.Node)) *NodeSaleDataGatewayWithTx_CreateNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(entity.Node))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateNode_Call) Return(_a0 error) *NodeSaleDataGatewayWithTx_CreateNode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateNode_Call) RunAndReturn(run func(context.Context, entity.Node) error) *NodeSaleDataGatewayWithTx_CreateNode_Call {
	_c.Call.Return(run)
	return _c
}

// CreateNodeSale provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) CreateNodeSale(ctx context.Context, arg entity.NodeSale) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for CreateNodeSale")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entity.NodeSale) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeSaleDataGatewayWithTx_CreateNodeSale_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateNodeSale'
type NodeSaleDataGatewayWithTx_CreateNodeSale_Call struct {
	*mock.Call
}

// CreateNodeSale is a helper method to define mock.On call
//   - ctx context.Context
//   - arg entity.NodeSale
func (_e *NodeSaleDataGatewayWithTx_Expecter) CreateNodeSale(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_CreateNodeSale_Call {
	return &NodeSaleDataGatewayWithTx_CreateNodeSale_Call{Call: _e.mock.On("CreateNodeSale", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_CreateNodeSale_Call) Run(run func(ctx context.Context, arg entity.NodeSale)) *NodeSaleDataGatewayWithTx_CreateNodeSale_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(entity.NodeSale))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateNodeSale_Call) Return(_a0 error) *NodeSaleDataGatewayWithTx_CreateNodeSale_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_CreateNodeSale_Call) RunAndReturn(run func(context.Context, entity.NodeSale) error) *NodeSaleDataGatewayWithTx_CreateNodeSale_Call {
	_c.Call.Return(run)
	return _c
}

// GetBlock provides a mock function with given fields: ctx, blockHeight
func (_m *NodeSaleDataGatewayWithTx) GetBlock(ctx context.Context, blockHeight int64) (*entity.Block, error) {
	ret := _m.Called(ctx, blockHeight)

	if len(ret) == 0 {
		panic("no return value specified for GetBlock")
	}

	var r0 *entity.Block
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*entity.Block, error)); ok {
		return rf(ctx, blockHeight)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *entity.Block); ok {
		r0 = rf(ctx, blockHeight)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Block)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, blockHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBlock'
type NodeSaleDataGatewayWithTx_GetBlock_Call struct {
	*mock.Call
}

// GetBlock is a helper method to define mock.On call
//   - ctx context.Context
//   - blockHeight int64
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetBlock(ctx interface{}, blockHeight interface{}) *NodeSaleDataGatewayWithTx_GetBlock_Call {
	return &NodeSaleDataGatewayWithTx_GetBlock_Call{Call: _e.mock.On("GetBlock", ctx, blockHeight)}
}

func (_c *NodeSaleDataGatewayWithTx_GetBlock_Call) Run(run func(ctx context.Context, blockHeight int64)) *NodeSaleDataGatewayWithTx_GetBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetBlock_Call) Return(_a0 *entity.Block, _a1 error) *NodeSaleDataGatewayWithTx_GetBlock_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetBlock_Call) RunAndReturn(run func(context.Context, int64) (*entity.Block, error)) *NodeSaleDataGatewayWithTx_GetBlock_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventsByWallet provides a mock function with given fields: ctx, walletAddress
func (_m *NodeSaleDataGatewayWithTx) GetEventsByWallet(ctx context.Context, walletAddress string) ([]entity.NodeSaleEvent, error) {
	ret := _m.Called(ctx, walletAddress)

	if len(ret) == 0 {
		panic("no return value specified for GetEventsByWallet")
	}

	var r0 []entity.NodeSaleEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]entity.NodeSaleEvent, error)); ok {
		return rf(ctx, walletAddress)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []entity.NodeSaleEvent); ok {
		r0 = rf(ctx, walletAddress)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.NodeSaleEvent)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, walletAddress)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetEventsByWallet_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventsByWallet'
type NodeSaleDataGatewayWithTx_GetEventsByWallet_Call struct {
	*mock.Call
}

// GetEventsByWallet is a helper method to define mock.On call
//   - ctx context.Context
//   - walletAddress string
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetEventsByWallet(ctx interface{}, walletAddress interface{}) *NodeSaleDataGatewayWithTx_GetEventsByWallet_Call {
	return &NodeSaleDataGatewayWithTx_GetEventsByWallet_Call{Call: _e.mock.On("GetEventsByWallet", ctx, walletAddress)}
}

func (_c *NodeSaleDataGatewayWithTx_GetEventsByWallet_Call) Run(run func(ctx context.Context, walletAddress string)) *NodeSaleDataGatewayWithTx_GetEventsByWallet_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetEventsByWallet_Call) Return(_a0 []entity.NodeSaleEvent, _a1 error) *NodeSaleDataGatewayWithTx_GetEventsByWallet_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetEventsByWallet_Call) RunAndReturn(run func(context.Context, string) ([]entity.NodeSaleEvent, error)) *NodeSaleDataGatewayWithTx_GetEventsByWallet_Call {
	_c.Call.Return(run)
	return _c
}

// GetLastProcessedBlock provides a mock function with given fields: ctx
func (_m *NodeSaleDataGatewayWithTx) GetLastProcessedBlock(ctx context.Context) (*entity.Block, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetLastProcessedBlock")
	}

	var r0 *entity.Block
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*entity.Block, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *entity.Block); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Block)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLastProcessedBlock'
type NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call struct {
	*mock.Call
}

// GetLastProcessedBlock is a helper method to define mock.On call
//   - ctx context.Context
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetLastProcessedBlock(ctx interface{}) *NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call {
	return &NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call{Call: _e.mock.On("GetLastProcessedBlock", ctx)}
}

func (_c *NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call) Run(run func(ctx context.Context)) *NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call) Return(_a0 *entity.Block, _a1 error) *NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call) RunAndReturn(run func(context.Context) (*entity.Block, error)) *NodeSaleDataGatewayWithTx_GetLastProcessedBlock_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodeCountByTierIndex provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) GetNodeCountByTierIndex(ctx context.Context, arg datagateway.GetNodeCountByTierIndexParams) ([]datagateway.GetNodeCountByTierIndexRow, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for GetNodeCountByTierIndex")
	}

	var r0 []datagateway.GetNodeCountByTierIndexRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodeCountByTierIndexParams) ([]datagateway.GetNodeCountByTierIndexRow, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodeCountByTierIndexParams) []datagateway.GetNodeCountByTierIndexRow); ok {
		r0 = rf(ctx, arg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]datagateway.GetNodeCountByTierIndexRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, datagateway.GetNodeCountByTierIndexParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodeCountByTierIndex'
type NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call struct {
	*mock.Call
}

// GetNodeCountByTierIndex is a helper method to define mock.On call
//   - ctx context.Context
//   - arg datagateway.GetNodeCountByTierIndexParams
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetNodeCountByTierIndex(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call {
	return &NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call{Call: _e.mock.On("GetNodeCountByTierIndex", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call) Run(run func(ctx context.Context, arg datagateway.GetNodeCountByTierIndexParams)) *NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datagateway.GetNodeCountByTierIndexParams))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call) Return(_a0 []datagateway.GetNodeCountByTierIndexRow, _a1 error) *NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call) RunAndReturn(run func(context.Context, datagateway.GetNodeCountByTierIndexParams) ([]datagateway.GetNodeCountByTierIndexRow, error)) *NodeSaleDataGatewayWithTx_GetNodeCountByTierIndex_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodeSale provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) GetNodeSale(ctx context.Context, arg datagateway.GetNodeSaleParams) ([]entity.NodeSale, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for GetNodeSale")
	}

	var r0 []entity.NodeSale
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodeSaleParams) ([]entity.NodeSale, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodeSaleParams) []entity.NodeSale); ok {
		r0 = rf(ctx, arg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.NodeSale)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, datagateway.GetNodeSaleParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetNodeSale_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodeSale'
type NodeSaleDataGatewayWithTx_GetNodeSale_Call struct {
	*mock.Call
}

// GetNodeSale is a helper method to define mock.On call
//   - ctx context.Context
//   - arg datagateway.GetNodeSaleParams
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetNodeSale(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_GetNodeSale_Call {
	return &NodeSaleDataGatewayWithTx_GetNodeSale_Call{Call: _e.mock.On("GetNodeSale", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_GetNodeSale_Call) Run(run func(ctx context.Context, arg datagateway.GetNodeSaleParams)) *NodeSaleDataGatewayWithTx_GetNodeSale_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datagateway.GetNodeSaleParams))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodeSale_Call) Return(_a0 []entity.NodeSale, _a1 error) *NodeSaleDataGatewayWithTx_GetNodeSale_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodeSale_Call) RunAndReturn(run func(context.Context, datagateway.GetNodeSaleParams) ([]entity.NodeSale, error)) *NodeSaleDataGatewayWithTx_GetNodeSale_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodesByIds provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) GetNodesByIds(ctx context.Context, arg datagateway.GetNodesByIdsParams) ([]entity.Node, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for GetNodesByIds")
	}

	var r0 []entity.Node
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodesByIdsParams) ([]entity.Node, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodesByIdsParams) []entity.Node); ok {
		r0 = rf(ctx, arg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.Node)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, datagateway.GetNodesByIdsParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetNodesByIds_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodesByIds'
type NodeSaleDataGatewayWithTx_GetNodesByIds_Call struct {
	*mock.Call
}

// GetNodesByIds is a helper method to define mock.On call
//   - ctx context.Context
//   - arg datagateway.GetNodesByIdsParams
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetNodesByIds(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_GetNodesByIds_Call {
	return &NodeSaleDataGatewayWithTx_GetNodesByIds_Call{Call: _e.mock.On("GetNodesByIds", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByIds_Call) Run(run func(ctx context.Context, arg datagateway.GetNodesByIdsParams)) *NodeSaleDataGatewayWithTx_GetNodesByIds_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datagateway.GetNodesByIdsParams))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByIds_Call) Return(_a0 []entity.Node, _a1 error) *NodeSaleDataGatewayWithTx_GetNodesByIds_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByIds_Call) RunAndReturn(run func(context.Context, datagateway.GetNodesByIdsParams) ([]entity.Node, error)) *NodeSaleDataGatewayWithTx_GetNodesByIds_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodesByOwner provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) GetNodesByOwner(ctx context.Context, arg datagateway.GetNodesByOwnerParams) ([]entity.Node, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for GetNodesByOwner")
	}

	var r0 []entity.Node
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodesByOwnerParams) ([]entity.Node, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodesByOwnerParams) []entity.Node); ok {
		r0 = rf(ctx, arg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.Node)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, datagateway.GetNodesByOwnerParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetNodesByOwner_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodesByOwner'
type NodeSaleDataGatewayWithTx_GetNodesByOwner_Call struct {
	*mock.Call
}

// GetNodesByOwner is a helper method to define mock.On call
//   - ctx context.Context
//   - arg datagateway.GetNodesByOwnerParams
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetNodesByOwner(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_GetNodesByOwner_Call {
	return &NodeSaleDataGatewayWithTx_GetNodesByOwner_Call{Call: _e.mock.On("GetNodesByOwner", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByOwner_Call) Run(run func(ctx context.Context, arg datagateway.GetNodesByOwnerParams)) *NodeSaleDataGatewayWithTx_GetNodesByOwner_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datagateway.GetNodesByOwnerParams))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByOwner_Call) Return(_a0 []entity.Node, _a1 error) *NodeSaleDataGatewayWithTx_GetNodesByOwner_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByOwner_Call) RunAndReturn(run func(context.Context, datagateway.GetNodesByOwnerParams) ([]entity.Node, error)) *NodeSaleDataGatewayWithTx_GetNodesByOwner_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodesByPubkey provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) GetNodesByPubkey(ctx context.Context, arg datagateway.GetNodesByPubkeyParams) ([]entity.Node, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for GetNodesByPubkey")
	}

	var r0 []entity.Node
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodesByPubkeyParams) ([]entity.Node, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.GetNodesByPubkeyParams) []entity.Node); ok {
		r0 = rf(ctx, arg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.Node)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, datagateway.GetNodesByPubkeyParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodesByPubkey'
type NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call struct {
	*mock.Call
}

// GetNodesByPubkey is a helper method to define mock.On call
//   - ctx context.Context
//   - arg datagateway.GetNodesByPubkeyParams
func (_e *NodeSaleDataGatewayWithTx_Expecter) GetNodesByPubkey(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call {
	return &NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call{Call: _e.mock.On("GetNodesByPubkey", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call) Run(run func(ctx context.Context, arg datagateway.GetNodesByPubkeyParams)) *NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datagateway.GetNodesByPubkeyParams))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call) Return(_a0 []entity.Node, _a1 error) *NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call) RunAndReturn(run func(context.Context, datagateway.GetNodesByPubkeyParams) ([]entity.Node, error)) *NodeSaleDataGatewayWithTx_GetNodesByPubkey_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveBlockFrom provides a mock function with given fields: ctx, fromBlock
func (_m *NodeSaleDataGatewayWithTx) RemoveBlockFrom(ctx context.Context, fromBlock int64) (int64, error) {
	ret := _m.Called(ctx, fromBlock)

	if len(ret) == 0 {
		panic("no return value specified for RemoveBlockFrom")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (int64, error)); ok {
		return rf(ctx, fromBlock)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) int64); ok {
		r0 = rf(ctx, fromBlock)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, fromBlock)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveBlockFrom'
type NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call struct {
	*mock.Call
}

// RemoveBlockFrom is a helper method to define mock.On call
//   - ctx context.Context
//   - fromBlock int64
func (_e *NodeSaleDataGatewayWithTx_Expecter) RemoveBlockFrom(ctx interface{}, fromBlock interface{}) *NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call {
	return &NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call{Call: _e.mock.On("RemoveBlockFrom", ctx, fromBlock)}
}

func (_c *NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call) Run(run func(ctx context.Context, fromBlock int64)) *NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call) Return(_a0 int64, _a1 error) *NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call) RunAndReturn(run func(context.Context, int64) (int64, error)) *NodeSaleDataGatewayWithTx_RemoveBlockFrom_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveEventsFromBlock provides a mock function with given fields: ctx, fromBlock
func (_m *NodeSaleDataGatewayWithTx) RemoveEventsFromBlock(ctx context.Context, fromBlock int64) (int64, error) {
	ret := _m.Called(ctx, fromBlock)

	if len(ret) == 0 {
		panic("no return value specified for RemoveEventsFromBlock")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (int64, error)); ok {
		return rf(ctx, fromBlock)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) int64); ok {
		r0 = rf(ctx, fromBlock)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, fromBlock)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveEventsFromBlock'
type NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call struct {
	*mock.Call
}

// RemoveEventsFromBlock is a helper method to define mock.On call
//   - ctx context.Context
//   - fromBlock int64
func (_e *NodeSaleDataGatewayWithTx_Expecter) RemoveEventsFromBlock(ctx interface{}, fromBlock interface{}) *NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call {
	return &NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call{Call: _e.mock.On("RemoveEventsFromBlock", ctx, fromBlock)}
}

func (_c *NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call) Run(run func(ctx context.Context, fromBlock int64)) *NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call) Return(_a0 int64, _a1 error) *NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call) RunAndReturn(run func(context.Context, int64) (int64, error)) *NodeSaleDataGatewayWithTx_RemoveEventsFromBlock_Call {
	_c.Call.Return(run)
	return _c
}

// Rollback provides a mock function with given fields: ctx
func (_m *NodeSaleDataGatewayWithTx) Rollback(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NodeSaleDataGatewayWithTx_Rollback_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rollback'
type NodeSaleDataGatewayWithTx_Rollback_Call struct {
	*mock.Call
}

// Rollback is a helper method to define mock.On call
//   - ctx context.Context
func (_e *NodeSaleDataGatewayWithTx_Expecter) Rollback(ctx interface{}) *NodeSaleDataGatewayWithTx_Rollback_Call {
	return &NodeSaleDataGatewayWithTx_Rollback_Call{Call: _e.mock.On("Rollback", ctx)}
}

func (_c *NodeSaleDataGatewayWithTx_Rollback_Call) Run(run func(ctx context.Context)) *NodeSaleDataGatewayWithTx_Rollback_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_Rollback_Call) Return(_a0 error) *NodeSaleDataGatewayWithTx_Rollback_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_Rollback_Call) RunAndReturn(run func(context.Context) error) *NodeSaleDataGatewayWithTx_Rollback_Call {
	_c.Call.Return(run)
	return _c
}

// SetDelegates provides a mock function with given fields: ctx, arg
func (_m *NodeSaleDataGatewayWithTx) SetDelegates(ctx context.Context, arg datagateway.SetDelegatesParams) (int64, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for SetDelegates")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.SetDelegatesParams) (int64, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datagateway.SetDelegatesParams) int64); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, datagateway.SetDelegatesParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NodeSaleDataGatewayWithTx_SetDelegates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetDelegates'
type NodeSaleDataGatewayWithTx_SetDelegates_Call struct {
	*mock.Call
}

// SetDelegates is a helper method to define mock.On call
//   - ctx context.Context
//   - arg datagateway.SetDelegatesParams
func (_e *NodeSaleDataGatewayWithTx_Expecter) SetDelegates(ctx interface{}, arg interface{}) *NodeSaleDataGatewayWithTx_SetDelegates_Call {
	return &NodeSaleDataGatewayWithTx_SetDelegates_Call{Call: _e.mock.On("SetDelegates", ctx, arg)}
}

func (_c *NodeSaleDataGatewayWithTx_SetDelegates_Call) Run(run func(ctx context.Context, arg datagateway.SetDelegatesParams)) *NodeSaleDataGatewayWithTx_SetDelegates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datagateway.SetDelegatesParams))
	})
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_SetDelegates_Call) Return(_a0 int64, _a1 error) *NodeSaleDataGatewayWithTx_SetDelegates_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NodeSaleDataGatewayWithTx_SetDelegates_Call) RunAndReturn(run func(context.Context, datagateway.SetDelegatesParams) (int64, error)) *NodeSaleDataGatewayWithTx_SetDelegates_Call {
	_c.Call.Return(run)
	return _c
}

// NewNodeSaleDataGatewayWithTx creates a new instance of NodeSaleDataGatewayWithTx. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNodeSaleDataGatewayWithTx(t interface {
	mock.TestingT
	Cleanup(func())
}) *NodeSaleDataGatewayWithTx {
	mock := &NodeSaleDataGatewayWithTx{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
