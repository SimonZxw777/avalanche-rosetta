package pchain

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

var (
	cChainId = "yH8D7ThNJkxmtkuv2jgBa4P1Rn3Qpr4pPr7QYNfcdoS6k6HWp"
	chainIDs = map[string]string{
		ids.Empty.String(): mapper.PChainNetworkIdentifier,
		cChainId:           mapper.CChainNetworkIdentifier,
	}
)

func TestMapInOperation(t *testing.T) {
	addValidatorTx, inputAccounts := buildValidatorTx()

	assert.Equal(t, 1, len(addValidatorTx.Ins))
	assert.Equal(t, 0, len(addValidatorTx.Outs))

	avaxIn := addValidatorTx.Ins[0]
	parser := NewTxParser(false, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaInOp, err := parser.insToOperations(9, OpAddValidator, []*avax.TransferableInput{avaxIn}, OpTypeInput)
	assert.Nil(t, err)

	assert.Equal(t, int64(9), rosettaInOp[0].OperationIdentifier.Index)
	assert.Equal(t, OpAddValidator, rosettaInOp[0].Type)
	assert.Equal(t, avaxIn.UTXOID.String(), rosettaInOp[0].CoinChange.CoinIdentifier.Identifier)
	assert.Equal(t, types.CoinSpent, rosettaInOp[0].CoinChange.CoinAction)
	assert.Equal(t, OpTypeInput, rosettaInOp[0].Metadata["type"])
	assert.Equal(t, OpAddValidator, rosettaInOp[0].Type)
	assert.Equal(t, types.String(mapper.StatusSuccess), rosettaInOp[0].Status)

	assert.Nil(t, rosettaInOp[0].Metadata["threshold"])
	assert.NotNil(t, rosettaInOp[0].Metadata["sig_indices"])
}

func TestMapOutOperation(t *testing.T) {
	addDelegatorTx, inputAccounts := buildAddDelegator()

	assert.Equal(t, 1, len(addDelegatorTx.Ins))
	assert.Equal(t, 1, len(addDelegatorTx.Outs))

	avaxOut := addDelegatorTx.Outs[0]

	parser := NewTxParser(true, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaOutOp, _, err := parser.outsToOperations(0, 0, OpAddDelegator, ids.Empty, []*avax.TransferableOutput{avaxOut}, OpTypeOutput, mapper.PChainNetworkIdentifier)
	assert.Nil(t, err)

	assert.Equal(t, int64(0), rosettaOutOp[0].OperationIdentifier.Index)
	assert.Equal(t, "P-fuji1gdkq8g208e3j4epyjmx65jglsw7vauh86l47ac", rosettaOutOp[0].Account.Address)
	assert.Equal(t, mapper.AtomicAvaxCurrency, rosettaOutOp[0].Amount.Currency)
	assert.Equal(t, "996649063", rosettaOutOp[0].Amount.Value)
	assert.Equal(t, OpTypeOutput, rosettaOutOp[0].Metadata["type"])
	assert.Nil(t, rosettaOutOp[0].Status)
	assert.Equal(t, OpAddDelegator, rosettaOutOp[0].Type)

	assert.NotNil(t, rosettaOutOp[0].Metadata["threshold"])
	assert.NotNil(t, rosettaOutOp[0].Metadata["locktime"])
	assert.Nil(t, rosettaOutOp[0].Metadata["sig_indices"])

}

func TestMapAddValidatorTx(t *testing.T) {
	addvalidatorTx, inputAccounts := buildValidatorTx()

	assert.Equal(t, 1, len(addvalidatorTx.Ins))
	assert.Equal(t, 0, len(addvalidatorTx.Outs))

	parser := NewTxParser(true, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaTransaction, err := parser.Parse(addvalidatorTx)
	assert.Nil(t, err)

	total := len(addvalidatorTx.Ins) + len(addvalidatorTx.Outs) + len(addvalidatorTx.Stake)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	cntTxType, cntInputMeta, cntOutputMeta, cntMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpAddValidator, OpTypeStakeOutput)

	assert.Equal(t, 2, cntTxType)
	assert.Equal(t, 1, cntInputMeta)
	assert.Equal(t, 0, cntOutputMeta)
	assert.Equal(t, 1, cntMetaType)

}

func TestMapAddDelegatorTx(t *testing.T) {
	addDelegatorTx, inputAccounts := buildAddDelegator()

	assert.Equal(t, 1, len(addDelegatorTx.Ins))
	assert.Equal(t, 1, len(addDelegatorTx.Outs))
	assert.Equal(t, 1, len(addDelegatorTx.Stake))

	parser := NewTxParser(true, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaTransaction, err := parser.Parse(addDelegatorTx)
	assert.Nil(t, err)

	total := len(addDelegatorTx.Ins) + len(addDelegatorTx.Outs) + len(addDelegatorTx.Stake)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	cntTxType, cntInputMeta, cntOutputMeta, cntMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpAddDelegator, OpTypeStakeOutput)

	assert.Equal(t, 3, cntTxType)
	assert.Equal(t, 1, cntInputMeta)
	assert.Equal(t, 1, cntOutputMeta)
	assert.Equal(t, 1, cntMetaType)

	assert.Equal(t, types.CoinSpent, rosettaTransaction.Operations[0].CoinChange.CoinAction)
	assert.Nil(t, rosettaTransaction.Operations[1].CoinChange)
	assert.Nil(t, rosettaTransaction.Operations[2].CoinChange)

	assert.Equal(t, addDelegatorTx.Ins[0].UTXOID.String(), rosettaTransaction.Operations[0].CoinChange.CoinIdentifier.Identifier)

	assert.Equal(t, int64(0), rosettaTransaction.Operations[0].OperationIdentifier.Index)
	assert.Equal(t, int64(1), rosettaTransaction.Operations[1].OperationIdentifier.Index)
	assert.Equal(t, int64(2), rosettaTransaction.Operations[2].OperationIdentifier.Index)

	assert.Equal(t, OpAddDelegator, rosettaTransaction.Operations[0].Type)
	assert.Equal(t, OpAddDelegator, rosettaTransaction.Operations[1].Type)
	assert.Equal(t, OpAddDelegator, rosettaTransaction.Operations[2].Type)

	assert.Equal(t, OpTypeInput, rosettaTransaction.Operations[0].Metadata["type"])
	assert.Equal(t, OpTypeOutput, rosettaTransaction.Operations[1].Metadata["type"])
	assert.Equal(t, OpTypeStakeOutput, rosettaTransaction.Operations[2].Metadata["type"])
}
func TestMapImportTx(t *testing.T) {
	importTx, inputAccounts := buildImport()

	assert.Equal(t, 0, len(importTx.Ins))
	assert.Equal(t, 2, len(importTx.Outs))
	assert.Equal(t, 1, len(importTx.ImportedInputs))

	parser := NewTxParser(true, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaTransaction, err := parser.Parse(importTx)
	assert.Nil(t, err)

	total := len(importTx.Ins) + len(importTx.Outs) + len(importTx.ImportedInputs) - 1 // - 1 for the multisig output
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	cntTxType, cntInputMeta, cntOutputMeta, cntMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpImportAvax, OpTypeImport)

	assert.Equal(t, 2, cntTxType)
	assert.Equal(t, 0, cntInputMeta)
	assert.Equal(t, 1, cntOutputMeta)
	assert.Equal(t, 1, cntMetaType)

	assert.Equal(t, types.CoinSpent, rosettaTransaction.Operations[0].CoinChange.CoinAction)
	assert.Nil(t, rosettaTransaction.Operations[1].CoinChange)
}

func TestNonConstructionMapExportTx(t *testing.T) {
	exportTx, inputAccounts := buildExport()

	assert.Equal(t, 1, len(exportTx.Ins))
	assert.Equal(t, 1, len(exportTx.Outs))
	assert.Equal(t, 1, len(exportTx.ExportedOutputs))

	parser := NewTxParser(false, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaTransaction, err := parser.Parse(exportTx)
	assert.Nil(t, err)

	total := len(exportTx.Ins) + len(exportTx.Outs)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	cntTxType, cntInputMeta, cntOutputMeta, cntMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpExportAvax, OpTypeExport)

	assert.Equal(t, 2, cntTxType)
	assert.Equal(t, 1, cntInputMeta)
	assert.Equal(t, 1, cntOutputMeta)
	assert.Equal(t, 0, cntMetaType)

	txType, ok := rosettaTransaction.Metadata[MetadataTxType].(string)
	assert.True(t, ok)
	assert.Equal(t, OpExportAvax, txType)

	// Verify that skipped output is exactly the same as normal export output
	skippedOuts, ok := rosettaTransaction.Metadata[MetadataSkippedOuts].([]*types.Operation)
	assert.True(t, ok)

	parser = NewTxParser(true, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaTransaction, err = parser.Parse(exportTx)
	out := rosettaTransaction.Operations[2]
	out.Status = types.String(mapper.StatusSuccess)
	out.CoinChange = skippedOuts[0].CoinChange
	assert.Equal(t, []*types.Operation{out}, skippedOuts)
}

func TestMapExportTx(t *testing.T) {
	exportTx, inputAccounts := buildExport()

	assert.Equal(t, 1, len(exportTx.Ins))
	assert.Equal(t, 1, len(exportTx.Outs))
	assert.Equal(t, 1, len(exportTx.ExportedOutputs))

	parser := NewTxParser(true, constants.FujiHRP, chainIDs, inputAccounts, nil)
	rosettaTransaction, err := parser.Parse(exportTx)
	assert.Nil(t, err)

	total := len(exportTx.Ins) + len(exportTx.Outs) + len(exportTx.ExportedOutputs)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	cntTxType, cntInputMeta, cntOutputMeta, cntMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpExportAvax, OpTypeExport)

	assert.Equal(t, 3, cntTxType)
	assert.Equal(t, 1, cntInputMeta)
	assert.Equal(t, 1, cntOutputMeta)
	assert.Equal(t, 1, cntMetaType)
}

func verifyRosettaTransaction(operations []*types.Operation, txType string, metaType string) (int, int, int, int) {
	cntOpInputMeta := 0
	cntOpOutputMeta := 0
	cntTxType := 0
	cntMetaType := 0

	for _, v := range operations {
		if v.Type == txType {
			cntTxType++
		}

		meta := &OperationMetadata{}
		_ = mapper.UnmarshalJSONMap(v.Metadata, meta)

		if meta.Type == OpTypeInput {
			cntOpInputMeta++
			continue
		}
		if meta.Type == OpTypeOutput {
			cntOpOutputMeta++
			continue
		}
		if meta.Type == metaType {
			cntMetaType++
			continue
		}
	}

	return cntTxType, cntOpInputMeta, cntOpOutputMeta, cntMetaType
}
