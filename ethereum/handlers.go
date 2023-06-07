package ethereum

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"math/big"
	"net/http"
	"strconv"
)

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

func getBlock(w http.ResponseWriter, r *http.Request, cfg EthereumConfig) {
	decoder := json.NewDecoder(r.Body)
	var req BlockRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	blockNumber, err := strconv.ParseInt(req.BlockNumber, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(cfg.IPCPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	block, err := client.BlockByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, tx := range block.Transactions() {
		fmt.Println(tx.Hash().Hex())
	}

	response := BlockResponse{
		Number:           block.Number().Int64(),
		Hash:             block.Hash().Hex(),
		ParentHash:       block.ParentHash().Hex(),
		Sha3Uncles:       block.UncleHash().Hex(),
		TransactionsRoot: block.TxHash().Hex(),
		StateRoot:        block.Root().Hex(),
		ReceiptsRoot:     block.ReceiptHash().Hex(),
		Miner:            block.Coinbase().Hex(),
		Difficulty:       block.Difficulty().Uint64(),
		TotalDifficulty:  block.Difficulty().Uint64(), // This field may need to be adjusted.
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		Timestamp:        block.Time(),
		ExtraData:        hex.EncodeToString(block.Extra()),
		MixHash:          block.MixDigest().Hex(),
		Nonce:            hex.EncodeToString(block.Header().Nonce[:]),
	}

	json.NewEncoder(w).Encode(response)
}

func getBlockRange(w http.ResponseWriter, r *http.Request, cfg EthereumConfig) {
	decoder := json.NewDecoder(r.Body)
	var req BlockRangeRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startBlockNumber, err := strconv.ParseInt(req.StartBlock, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	endBlockNumber, err := strconv.ParseInt(req.EndBlock, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(cfg.IPCPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blockResponses := make(chan BlockResponse, endBlockNumber-startBlockNumber+1)
	errors := make(chan error, endBlockNumber-startBlockNumber+1)
	defer close(blockResponses)
	defer close(errors)

	for i := startBlockNumber; i <= endBlockNumber; i++ {
		go getBlockAsync(client, big.NewInt(i), blockResponses, errors)
	}

	blocks := make([]BlockResponse, 0, endBlockNumber-startBlockNumber+1)
	for i := startBlockNumber; i <= endBlockNumber; i++ {
		select {
		case block := <-blockResponses:
			blocks = append(blocks, block)
		case err := <-errors:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(blocks)
}

func getBlockAsync(client *ethclient.Client, blockNumber *big.Int, blockResponses chan<- BlockResponse, errors chan<- error) {
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		errors <- err
		return
	}

	blockResponse := BlockResponse{
		Number:           block.Number().Int64(),
		Hash:             block.Hash().Hex(),
		ParentHash:       block.ParentHash().Hex(),
		Sha3Uncles:       block.UncleHash().Hex(),
		TransactionsRoot: block.TxHash().Hex(),
		StateRoot:        block.Root().Hex(),
		ReceiptsRoot:     block.ReceiptHash().Hex(),
		Miner:            block.Coinbase().Hex(),
		Difficulty:       block.Difficulty().Uint64(),
		TotalDifficulty:  block.Difficulty().Uint64(), // This field may need to be adjusted.
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		Timestamp:        block.Time(),
		ExtraData:        hex.EncodeToString(block.Extra()),
		MixHash:          block.MixDigest().Hex(),
		Nonce:            hex.EncodeToString(block.Header().Nonce[:]),
	}

	blockResponses <- blockResponse
}

func RegisterHandlers(router *mux.Router, cfg EthereumConfig) {
	router.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		getBlock(w, r, cfg)
	}).Methods("POST")
	router.HandleFunc("/blockrange", func(w http.ResponseWriter, r *http.Request) {
		getBlockRange(w, r, cfg)
	}).Methods("POST")
}
