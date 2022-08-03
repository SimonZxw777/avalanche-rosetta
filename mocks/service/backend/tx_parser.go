// Code generated by mockery v2.12.3. DO NOT EDIT.

package chain

import (
	common "github.com/ava-labs/avalanche-rosetta/service/backend/common"
	mock "github.com/stretchr/testify/mock"

	types "github.com/coinbase/rosetta-sdk-go/types"
)

// TxParser is an autogenerated mock type for the TxParser type
type TxParser struct {
	mock.Mock
}

// ParseTx provides a mock function with given fields: tx, inputAddresses
func (_m *TxParser) ParseTx(tx *common.RosettaTx, inputAddresses map[string]*types.AccountIdentifier) ([]*types.Operation, error) {
	ret := _m.Called(tx, inputAddresses)

	var r0 []*types.Operation
	if rf, ok := ret.Get(0).(func(*common.RosettaTx, map[string]*types.AccountIdentifier) []*types.Operation); ok {
		r0 = rf(tx, inputAddresses)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*common.RosettaTx, map[string]*types.AccountIdentifier) error); ok {
		r1 = rf(tx, inputAddresses)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewTxParserT interface {
	mock.TestingT
	Cleanup(func())
}

// NewTxParser creates a new instance of TxParser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTxParser(t NewTxParserT) *TxParser {
	mock := &TxParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
