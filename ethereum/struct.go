package ethereum

import "github.com/ethereum/go-ethereum/ethclient"

type EthereumConfig struct {
	IPCPath string `json:"ipcpath"`
}

type BlockRequest struct {
	BlockNumber string `json:"blocknumber"`
}

type BlockResponse struct {
	Number           int64
	Hash             string
	ParentHash       string
	Sha3Uncles       string
	LogsBloom        interface{}
	TransactionsRoot string
	StateRoot        string
	ReceiptsRoot     string
	Miner            string
	Difficulty       uint64
	TotalDifficulty  uint64
	Size             interface{}
	GasLimit         uint64
	GasUsed          uint64
	Timestamp        uint64
	ExtraData        string
	MixHash          string
	Nonce            string
}

type BlockRangeRequest struct {
	StartBlock string `json:"startblock"`
	EndBlock   string `json:"endblock"`
}

type TransactionRequest struct {
	TxID string `json:"txid"`
}

type TransactionResponse struct {
	BlockHash        string
	BlockNumber      string
	TransactionIndex uint
	From             string
	To               string
	Value            string
	GasPrice         string
	Gas              uint64
	Input            string
	Nonce            uint64
	Hash             string
}

type BalanceRequest struct {
	Address string `json:"address"`
}

type BalanceResponse struct {
	Balance string `json:"balance"`
}

type MultiBalanceRequest struct {
	Addresses []string `json:"addresses"`
}

type MultiBalanceResponse struct {
	Balances map[string]string `json:"balances"`
}

type BlockRange struct {
	StartBlock int64 `json:"startBlock"`
	EndBlock   int64 `json:"endBlock"`
}

type Config struct {
	Client *ethclient.Client
	// diğer alanlar...
}

type GetAddressRangeRequest struct {
	StartBlock string `json:"startBlock"`
	EndBlock   string `json:"endBlock"`
}

type AddressRangeRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type AddressBalance struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

type AddressListResponse struct {
	Addresses []AddressBalance `json:"addresses"`
}

const (
	blockBatchSize = 500000
	oneEthInWei    = 1e18
)
