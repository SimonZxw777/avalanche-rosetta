package service

import (
	"context"
	"math/big"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	corethTypes "github.com/ava-labs/coreth/core/types"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/coinbase/rosetta-sdk-go/utils"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/mapper"
)

type BlockBackend interface {
	ShouldHandleRequest(req interface{}) bool
	Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error)
	BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error)
}

// BlockService implements the /block/* endpoints
type BlockService struct {
	config *Config
	client client.Client

	genesisBlock  *types.Block
	pChainBackend BlockBackend
	pChainBlockID *ids.ID
}

// NewBlockService returns a new block servicer
func NewBlockService(config *Config, c client.Client, pChainBackend BlockBackend) server.BlockAPIServicer {
	return &BlockService{
		config:        config,
		client:        c,
		genesisBlock:  makeGenesisBlock(config.GenesisBlockHash),
		pChainBackend: pChainBackend,
	}
}

// Block implements the /block endpoint
func (s *BlockService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
	if s.config.IsOfflineMode() {
		return nil, ErrUnavailableOffline
	}

	if request.BlockIdentifier == nil {
		return nil, ErrBlockInvalidInput
	}
	if request.BlockIdentifier.Hash == nil && request.BlockIdentifier.Index == nil {
		return nil, ErrBlockInvalidInput
	}

	if s.pChainBackend.ShouldHandleRequest(request) {
		return s.pChainBackend.Block(ctx, request)
	}

	if s.isGenesisBlockRequest(request.BlockIdentifier) {
		return &types.BlockResponse{
			Block: s.genesisBlock,
		}, nil
	}

	var (
		blockIdentifier       *types.BlockIdentifier
		parentBlockIdentifier *types.BlockIdentifier
		block                 *corethTypes.Block
		err                   error
	)

	if hash := request.BlockIdentifier.Hash; hash != nil {
		block, err = s.client.BlockByHash(ctx, ethcommon.HexToHash(*hash))
	} else if index := request.BlockIdentifier.Index; block == nil && index != nil {
		block, err = s.client.BlockByNumber(ctx, big.NewInt(*index))
	}
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrBlockNotFound
		}
		return nil, WrapError(ErrClientError, err)
	}

	blockIdentifier = &types.BlockIdentifier{
		Index: block.Number().Int64(),
		Hash:  block.Hash().String(),
	}

	if block.ParentHash().String() != s.config.GenesisBlockHash {
		parentBlock, err := s.client.HeaderByHash(ctx, block.ParentHash())
		if err != nil {
			return nil, WrapError(ErrClientError, err)
		}

		parentBlockIdentifier = &types.BlockIdentifier{
			Index: parentBlock.Number.Int64(),
			Hash:  parentBlock.Hash().String(),
		}
	} else {
		parentBlockIdentifier = s.genesisBlock.BlockIdentifier
	}

	transactions, terr := s.fetchTransactions(ctx, block)
	if terr != nil {
		return nil, terr
	}

	if s.pChainBlockID == nil {
		id, err := s.client.GetBlockchainID(ctx, mapper.PChainNetworkIdentifier)
		if err != nil {
			return nil, WrapError(ErrInternalError, err)
		}
		s.pChainBlockID = &id
	}
	crosstx, terr := s.parseCrossChainTransactions(block, request.NetworkIdentifier)
	if terr != nil {
		return nil, terr
	}

	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       blockIdentifier,
			ParentBlockIdentifier: parentBlockIdentifier,
			Timestamp:             int64(block.Time() * utils.MillisecondsInSecond),
			Transactions:          append(transactions, crosstx...),
			Metadata:              mapper.BlockMetadata(block),
		},
	}, nil
}

// BlockTransaction implements the /block/transaction endpoint.
func (s *BlockService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	if s.config.IsOfflineMode() {
		return nil, ErrUnavailableOffline
	}

	if request.BlockIdentifier == nil {
		return nil, WrapError(ErrInvalidInput, "block identifier is not provided")
	}

	if s.pChainBackend.ShouldHandleRequest(request) {
		return s.pChainBackend.BlockTransaction(ctx, request)
	}

	header, err := s.client.HeaderByHash(ctx, ethcommon.HexToHash(request.BlockIdentifier.Hash))
	if err != nil {
		return nil, WrapError(ErrClientError, err)
	}

	hash := ethcommon.HexToHash(request.TransactionIdentifier.Hash)
	tx, pending, err := s.client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, WrapError(ErrClientError, err)
	}
	if pending {
		return nil, nil
	}

	trace, flattened, err := s.client.TraceTransaction(ctx, tx.Hash().String())
	if err != nil {
		return nil, WrapError(ErrClientError, err)
	}

	transaction, terr := s.fetchTransaction(ctx, tx, header, trace, flattened)
	if terr != nil {
		return nil, terr
	}

	return &types.BlockTransactionResponse{
		Transaction: transaction,
	}, nil
}

func (s *BlockService) fetchTransactions(
	ctx context.Context,
	block *corethTypes.Block,
) ([]*types.Transaction, *types.Error) {
	transactions := []*types.Transaction{}

	trace, flattened, err := s.client.TraceBlockByHash(ctx, block.Hash().String())
	if err != nil {
		return nil, WrapError(ErrClientError, err)
	}

	for i, tx := range block.Transactions() {
		transaction, terr := s.fetchTransaction(ctx, tx, block.Header(), trace[i], flattened[i])
		if terr != nil {
			return nil, terr
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (s *BlockService) fetchTransaction(
	ctx context.Context,
	tx *corethTypes.Transaction,
	header *corethTypes.Header,
	trace *client.Call,
	flattened []*client.FlatCall,
) (*types.Transaction, *types.Error) {
	msg, err := tx.AsMessage(s.config.Signer(), header.BaseFee)
	if err != nil {
		return nil, WrapError(ErrClientError, err)
	}

	receipt, err := s.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, WrapError(ErrClientError, err)
	}

	transaction, err := mapper.Transaction(header, tx, &msg, receipt, trace, flattened, s.client, s.config.IsAnalyticsMode(), s.config.TokenWhiteList, s.config.IndexUnknownTokens)
	if err != nil {
		return nil, WrapError(ErrInternalError, err)
	}

	return transaction, nil
}

func (s *BlockService) parseCrossChainTransactions(
	block *corethTypes.Block,
	networkIdentifier *types.NetworkIdentifier,
) ([]*types.Transaction, *types.Error) {
	result := []*types.Transaction{}

	crossTxs, err := mapper.CrossChainTransactions(s.config.AvaxAssetID, block, s.config.AP5Activation, networkIdentifier, s.pChainBlockID)
	if err != nil {
		return nil, WrapError(ErrInternalError, err)
	}

	for _, tx := range crossTxs {
		// Skip empty import/export transactions
		if len(tx.Operations) == 0 {
			continue
		}

		result = append(result, tx)
	}

	return result, nil
}

func (s *BlockService) isGenesisBlockRequest(id *types.PartialBlockIdentifier) bool {
	if number := id.Index; number != nil {
		return *number == s.genesisBlock.BlockIdentifier.Index
	}
	if hash := id.Hash; hash != nil {
		return *hash == s.genesisBlock.BlockIdentifier.Hash
	}
	return false
}
