package cchainatomictx

import (
	"context"
	"errors"
	"strconv"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/utils/math"

	"github.com/ava-labs/avalanche-rosetta/service/backend/common"

	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
)

func (b *Backend) AccountBalance(ctx context.Context, req *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	if req.AccountIdentifier == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "account identifier is not provided")
	}
	coins, wrappedErr := b.getAccountCoins(ctx, req.AccountIdentifier.Address)
	if wrappedErr != nil {
		return nil, wrappedErr
	}

	var balanceValue uint64

	for _, coin := range coins {
		amountValue, err := types.AmountValue(coin.Amount)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "unable to extract amount from UTXO")
		}

		balanceValue, err = math.Add64(balanceValue, amountValue.Uint64())
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "overflow while calculating balance")
		}
	}

	return &types.AccountBalanceResponse{
		//TODO: return block identifier once AvalancheGo exposes an API for it
		//BlockIdentifier: ...
		Balances: []*types.Amount{
			{
				Value:    strconv.FormatUint(balanceValue, 10),
				Currency: mapper.AvaxCurrency,
			},
		},
	}, nil
}

func (b *Backend) AccountCoins(ctx context.Context, req *types.AccountCoinsRequest) (*types.AccountCoinsResponse, *types.Error) {
	if req.AccountIdentifier == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "account identifier is not provided")
	}
	coins, wrappedErr := b.getAccountCoins(ctx, req.AccountIdentifier.Address)
	if wrappedErr != nil {
		return nil, wrappedErr
	}

	return &types.AccountCoinsResponse{
		//TODO: return block identifier once AvalancheGo exposes an API for it
		// BlockIdentifier: ...
		Coins: common.SortUnique(coins),
	}, nil
}

func (b *Backend) getAccountCoins(ctx context.Context, address string) ([]*types.Coin, *types.Error) {
	var coins []*types.Coin
	sourceChains := []string{
		mapper.PChainNetworkIdentifier,
		mapper.XChainNetworkIdentifier,
	}

	for _, chain := range sourceChains {
		chainCoins, wrappedErr := b.fetchCoinsFromChain(ctx, address, chain)
		if wrappedErr != nil {
			return nil, wrappedErr
		}
		coins = append(coins, chainCoins...)
	}

	return coins, nil
}

func (b *Backend) fetchCoinsFromChain(ctx context.Context, address string, sourceChain string) ([]*types.Coin, *types.Error) {
	var coins []*types.Coin

	// Used for pagination
	var lastUtxoIndex api.Index

	for {

		// GetUTXOs controlled by addr
		utxos, newUtxoIndex, err := b.cClient.GetAtomicUTXOs(ctx, []string{address}, sourceChain, b.getUTXOsPageSize, lastUtxoIndex.Address, lastUtxoIndex.UTXO)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "unable to get UTXOs")
		}

		// convert raw UTXO bytes to Rosetta Coins
		coinsPage, err := b.processUtxos(sourceChain, utxos)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}

		coins = append(coins, coinsPage...)

		// Fetch next page only if there may be more UTXOs
		if len(utxos) < int(b.getUTXOsPageSize) {
			break
		}

		lastUtxoIndex = newUtxoIndex
	}

	return coins, nil
}

func (b *Backend) processUtxos(sourceChain string, utxos [][]byte) ([]*types.Coin, error) {
	var coins []*types.Coin
	for _, utxoBytes := range utxos {
		utxo := avax.UTXO{}
		_, err := platformvm.Codec.Unmarshal(utxoBytes, &utxo)
		if err != nil {
			return nil, errors.New("unable to parse UTXO")
		}

		transferableOut, ok := utxo.Out.(avax.TransferableOut)
		if !ok {
			return nil, errors.New("unable to get UTXO output")
		}

		coin := &types.Coin{
			CoinIdentifier: &types.CoinIdentifier{Identifier: utxo.UTXOID.String()},
			Amount: &types.Amount{
				Value:    strconv.FormatUint(transferableOut.Amount(), 10),
				Currency: mapper.AvaxCurrency,
				Metadata: map[string]interface{}{
					"source_chain": sourceChain,
				},
			},
		}
		coins = append(coins, coin)
	}
	return coins, nil
}