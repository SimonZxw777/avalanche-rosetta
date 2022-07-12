package p

import (
	"context"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func (c *Backend) NetworkIdentifier() *types.NetworkIdentifier {
	return c.networkIdentifier
}

func (c *Backend) NetworkStatus(ctx context.Context, req *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	// Fetch peers
	infoPeers, err := c.pClient.Peers(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}
	peers := mapper.Peers(infoPeers)

	// Check if network is bootstrapped
	ready, err := c.pClient.IsBootstrapped(ctx, mapper.PChainNetworkIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	if !ready {
		return &types.NetworkStatusResponse{
			CurrentBlockIdentifier: c.genesisBlockIdentifier,
			CurrentBlockTimestamp:  c.genesisBlock.Timestamp,
			GenesisBlockIdentifier: c.genesisBlockIdentifier,
			SyncStatus:             mapper.StageBootstrap,
			Peers:                  peers,
		}, nil
	}

	// Current block height
	currentBlock, err := c.indexerParser.ParseCurrentBlock(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	return &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: int64(currentBlock.Height),
			Hash:  currentBlock.BlockID.String(),
		},
		CurrentBlockTimestamp:  mapper.UnixToUnixMilli(currentBlock.Timestamp),
		GenesisBlockIdentifier: c.genesisBlockIdentifier,
		SyncStatus:             mapper.StageSynced,
		Peers:                  peers,
	}, nil
}

func (c *Backend) NetworkOptions(ctx context.Context, request *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion:    types.RosettaAPIVersion,
			NodeVersion:       service.NodeVersion,
			MiddlewareVersion: types.String(service.MiddlewareVersion),
		},
		Allow: &types.Allow{
			OperationStatuses:       mapper.OperationStatuses,
			OperationTypes:          mapper.PChainOperationTypes,
			CallMethods:             mapper.PChainCallMethods,
			Errors:                  service.Errors,
			HistoricalBalanceLookup: false,
		},
	}, nil
}
