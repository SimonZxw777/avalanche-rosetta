// Code generated by mockery v2.12.3. DO NOT EDIT.

package chain

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/coinbase/rosetta-sdk-go/types"
)

// BlockBackend is an autogenerated mock type for the BlockBackend type
type BlockBackend struct {
	mock.Mock
}

// Block provides a mock function with given fields: ctx, request
func (_m *BlockBackend) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	ret := _m.Called(ctx, request)

	var r0 *types.BlockResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.BlockRequest) *types.BlockResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockResponse)
		}
	}

	var r1 *types.Error
	if rf, ok := ret.Get(1).(func(context.Context, *types.BlockRequest) *types.Error); ok {
		r1 = rf(ctx, request)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*types.Error)
		}
	}

	return r0, r1
}

// BlockTransaction provides a mock function with given fields: ctx, request
func (_m *BlockBackend) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	ret := _m.Called(ctx, request)

	var r0 *types.BlockTransactionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.BlockTransactionRequest) *types.BlockTransactionResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockTransactionResponse)
		}
	}

	var r1 *types.Error
	if rf, ok := ret.Get(1).(func(context.Context, *types.BlockTransactionRequest) *types.Error); ok {
		r1 = rf(ctx, request)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*types.Error)
		}
	}

	return r0, r1
}

type NewBlockBackendT interface {
	mock.TestingT
	Cleanup(func())
}

// NewBlockBackend creates a new instance of BlockBackend. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBlockBackend(t NewBlockBackendT) *BlockBackend {
	mock := &BlockBackend{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}