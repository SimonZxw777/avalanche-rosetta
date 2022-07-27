package pchain

import "github.com/ava-labs/avalanchego/ids"

const (
	OpImportAvax      = "IMPORT_AVAX"
	OpExportAvax      = "EXPORT_AVAX"
	OpAddValidator    = "ADD_VALIDATOR"
	OpAddDelegator    = "ADD_DELEGATOR"
	OpRewardValidator = "REWARD_VALIDATOR"
	OpCreateChain     = "CREATE_CHAIN"

	OpTypeImport      = "IMPORT"
	OpTypeExport      = "EXPORT"
	OpTypeInput       = "INPUT"
	OpTypeOutput      = "OUTPUT"
	OpTypeStakeOutput = "STAKE"

	MetadataOpType      = "type"
	MetadataStakingTxId = "staking_tx"
	MetadataSubnetID    = "subnet_id"
	MetadataChainName   = "chain_name"
	MetadataVMID        = "vmid"
	MetadataMemo        = "memo"
	MetadataMessage     = "message"
)

var (
	OperationTypes = []string{
		OpImportAvax,
		OpExportAvax,
		OpAddValidator,
		OpAddDelegator,
		OpRewardValidator,
	}
	CallMethods = []string{}
)

type OperationMetadata struct {
	Type       string   `json:"type"`
	SigIndices []uint32 `json:"sig_indices,omitempty"`
	Locktime   uint64   `json:"locktime"`
	Threshold  uint32   `json:"threshold,omitempty"`
}

type ImportExportOptions struct {
	SourceChain      string `json:"source_chain"`
	DestinationChain string `json:"destination_chain"`
}

type StakingOptions struct {
	NodeID          string   `json:"node_id"`
	Start           uint64   `json:"start"`
	End             uint64   `json:"end"`
	Wght            uint64   `json:"weight"`
	Shares          uint32   `json:"shares"`
	Memo            string   `json:"memo"`
	Locktime        uint64   `json:"locktime"`
	Threshold       uint32   `json:"threshold"`
	RewardAddresses []string `json:"reward_addresses"`
}

type Metadata struct {
	NetworkID    uint32 `json:"network_id"`
	BlockchainID ids.ID `json:"blockchain_id"`
	*ImportMetadata
	*ExportMetadata
	*StakingMetadata
}

type ImportMetadata struct {
	SourceChainID ids.ID `json:"source_chain_id"`
}

type ExportMetadata struct {
	DestinationChain   string `json:"destination_chain"`
	DestinationChainID ids.ID `json:"destination_chain_id"`
}

type StakingMetadata struct {
	NodeID          string   `json:"node_id"`
	RewardAddresses []string `json:"reward_addresses"`
	Start           uint64   `json:"start"`
	End             uint64   `json:"end"`
	Wght            uint64   `json:"weight"`
	Shares          uint32   `json:"shares"`
	Locktime        uint64   `json:"locktime"`
	Threshold       uint32   `json:"threshold"`
	Memo            string   `json:"memo"`
}
