package p

import (
	"context"
	"errors"
	"fmt"
	"github.com/ava-labs/avalanche-rosetta/service/chain/common"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"
	"log"
	"reflect"

	mapper "github.com/ava-labs/avalanche-rosetta/mapper/p"
	"github.com/ava-labs/avalanche-rosetta/service"
)

// Block implements the /block endpoint
func (b *Backend) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	if b.isGenesisBlockRequest(request.BlockIdentifier) {
		return b.makeGenesisBlock(), nil
	}

	var blockIndex int64
	var hash string

	if request.BlockIdentifier.Index != nil {
		blockIndex = *request.BlockIdentifier.Index
	}

	if request.BlockIdentifier.Hash != nil {
		hash = *request.BlockIdentifier.Hash
	}

	block, blockHash, err := b.getBlock(ctx, blockIndex, hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var txs []*platformvm.Tx
	switch v := block.(type) {
	case *platformvm.AbortBlock, *platformvm.CommitBlock, *platformvm.AtomicBlock:
	case *platformvm.ProposalBlock:
		txs = append(txs, &v.Tx)
	case *platformvm.StandardBlock:
		txs = append(txs, v.Txs...)
	default:
		log.Printf("unknown %s", reflect.TypeOf(v))
	}

	transactions := []*types.Transaction{}
	for _, tx := range txs {
		err := common.InitializeTx(b.codecVersion, b.codec, *tx)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "tx initalize error")
		}

		t, err := mapper.Transaction(tx.UnsignedTx)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
		transactions = append(transactions, t)
	}

	resp := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Height()),
				Hash:  blockHash,
			},
			ParentBlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Height()) - 1,
				Hash:  block.Parent().String(),
			},
			//TODO: Find a way to get block timestamp. The following causes panic as there is no vm defined for the block
			//Timestamp:    block.Timestamp().UnixMilli(),
			Transactions: transactions,
		},
	}

	return resp, nil
}

// BlockTransaction implements the /block/transaction endpoint.
func (b *Backend) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	id, err := ids.FromString(request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	block, _, err := b.getBlock(ctx, request.BlockIdentifier.Index, request.BlockIdentifier.Hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var txs []*platformvm.Tx
	switch typedBlock := block.(type) {
	case *platformvm.StandardBlock:
		txs = append(txs, typedBlock.Txs...)
	case *platformvm.ProposalBlock:
		txs = append(txs, &typedBlock.Tx)
	}

	for _, tx := range txs {
		err := common.InitializeTx(b.codecVersion, b.codec, *tx)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "tx initalize error")
		}

		if tx.ID() != id {
			continue
		}

		t, err := mapper.Transaction(tx.UnsignedTx)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}

		resp := &types.BlockTransactionResponse{
			Transaction: t,
		}
		return resp, nil
	}

	return nil, service.ErrTransactionNotFound
}

func (b *Backend) getBlock(ctx context.Context, index int64, hash string) (platformvm.Block, string, error) {
	var blockId ids.ID
	var err error

	if index <= 0 || hash == "" {
		return nil, "", errors.New("a positive block index, a block hash or both must be specified")
	}

	// Extract block id from hash parameter if it is non-empty, or from index if stated
	if hash != "" {
		blockId, err = ids.FromString(hash)
		if err != nil {
			return nil, "", err
		}
	} else if index > 0 {
		blockIndex := uint64(index)
		parsedBlock, err := b.indexerParser.ParseBlockAtIndex(ctx, blockIndex)
		if err != nil {
			return nil, "", err
		}

		blockId = parsedBlock.BlockID
	}

	// Get the block bytes
	var blockBytes []byte
	blockBytes, err = b.pClient.GetBlock(ctx, blockId)
	if err != nil {
		return nil, "", err
	}

	// Unmarshal the block
	var block platformvm.Block
	if _, err := platformvm.Codec.Unmarshal(blockBytes, &block); err != nil {
		return nil, "", err
	}

	// Verify block height matches specified height - if there is one
	if index > 0 && block.Height() != uint64(index) {
		return nil, "", fmt.Errorf("requested block index: %d, found: %d for block %s", index, block.Height(), blockId.String())
	}

	return block, blockId.String(), nil
}

func (b *Backend) isGenesisBlockRequest(id *types.PartialBlockIdentifier) bool {
	if number := id.Index; number != nil {
		return *number == b.genesisBlockIdentifier.Index
	}
	if hash := id.Hash; hash != nil {
		return *hash == b.genesisBlockIdentifier.Hash
	}
	return false
}
