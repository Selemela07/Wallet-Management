package ethereum

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"sync"
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

func RegisterHandlers(router *mux.Router, cfg EthereumConfig) {
	router.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		getBlock(w, r, cfg)
	}).Methods("POST")

	router.HandleFunc("/transaction", func(w http.ResponseWriter, r *http.Request) {
		getTransaction(w, r, cfg)
	}).Methods("POST")

	router.HandleFunc("/transaction", func(w http.ResponseWriter, r *http.Request) {
		getTransaction(w, r, cfg)
	}).Methods("POST")

	router.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		getBalance(w, r, cfg)
	}).Methods("POST")

	router.HandleFunc("/multibalance", func(w http.ResponseWriter, r *http.Request) {
		getMultiBalance(w, r, cfg)
	}).Methods("POST")

	router.HandleFunc("/uniqueaddresses", func(w http.ResponseWriter, r *http.Request) {
		getAllTransactionAddressesHandler(w, r, cfg)
	}).Methods("POST")
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

func getTransaction(w http.ResponseWriter, r *http.Request, cfg EthereumConfig) {
	decoder := json.NewDecoder(r.Body)
	var req TransactionRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(cfg.IPCPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	txHash := common.HexToHash(req.TxID)
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isPending {
		http.Error(w, "Transaction is still pending", http.StatusBadRequest)
		return
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signer := types.NewEIP155Signer(chainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TransactionResponse{
		BlockHash:        receipt.BlockHash.Hex(),
		BlockNumber:      receipt.BlockNumber.String(),
		TransactionIndex: receipt.TransactionIndex,
		From:             from.Hex(),
		To:               tx.To().Hex(),
		Value:            tx.Value().String(),
		GasPrice:         tx.GasPrice().String(),
		Gas:              tx.Gas(),
		Input:            hex.EncodeToString(tx.Data()),
		Nonce:            tx.Nonce(),
		Hash:             tx.Hash().Hex(),
	}

	json.NewEncoder(w).Encode(response)
}

func getBalance(w http.ResponseWriter, r *http.Request, cfg EthereumConfig) {
	decoder := json.NewDecoder(r.Body)
	var req BalanceRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(cfg.IPCPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	address := common.HexToAddress(req.Address)
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	etherBalance := weiToEther(balance)

	response := BalanceResponse{
		Balance: etherBalance.Text('f', 18), // 18 decimal places for Ether
	}

	json.NewEncoder(w).Encode(response)
}

func getMultiBalance(w http.ResponseWriter, r *http.Request, cfg EthereumConfig) {
	decoder := json.NewDecoder(r.Body)
	var req MultiBalanceRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(cfg.IPCPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balances := make(map[string]string)

	for _, addr := range req.Addresses {
		address := common.HexToAddress(addr)
		balance, err := client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		etherBalance := weiToEther(balance)
		balances[addr] = etherBalance.Text('f', 18)
	}

	response := MultiBalanceResponse{
		Balances: balances,
	}

	json.NewEncoder(w).Encode(response)
}

func getAllTransactionAddressesHandler(w http.ResponseWriter, r *http.Request, cfg EthereumConfig) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddressRangeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	start, ok := new(big.Int).SetString(req.Start, 10)
	if !ok {
		http.Error(w, "Invalid start block number", http.StatusBadRequest)
		return
	}
	end, ok := new(big.Int).SetString(req.End, 10)
	if !ok {
		http.Error(w, "Invalid end block number", http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(cfg.IPCPath)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	addresses, err := getAllTransactionAddresses(client, start, end)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching addresses: %v", err), http.StatusInternalServerError)
		return
	}

	response := AddressListResponse{
		Addresses: addresses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getAllTransactionAddresses(client *ethclient.Client, startBlock, endBlock *big.Int) ([]AddressBalance, error) {
	addressMap := make(map[string]struct{})
	balances := make([]AddressBalance, 0)

	blocks := make(chan *big.Int, 100)
	go func() {
		defer close(blocks)
		for blockNumber := new(big.Int).Set(endBlock); blockNumber.Cmp(startBlock) >= 0; blockNumber.Sub(blockNumber, big.NewInt(1)) {
			blocks <- blockNumber
		}
	}()

	addressCh := make(chan string, 100)
	go func() {
		defer close(addressCh)
		var wg sync.WaitGroup
		for blockNumber := range blocks {
			wg.Add(1)
			go func(blockNumber *big.Int) {
				defer wg.Done()
				block, err := client.BlockByNumber(context.Background(), blockNumber)
				if err != nil {
					log.Printf("Error fetching block %s: %v", blockNumber, err)
					return
				}
				for _, tx := range block.Transactions() {
					to := tx.To()
					if to != nil {
						addressCh <- to.Hex()
					}
				}
			}(blockNumber)
		}
		wg.Wait()
	}()

	for addr := range addressCh {
		addressMap[addr] = struct{}{}
	}

	oneEthInWei := big.NewInt(1e18)
	for addr := range addressMap {
		address := common.HexToAddress(addr)
		balance, err := client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			return nil, err
		}

		if balance.Cmp(oneEthInWei) > 0 {
			balanceInEth := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(oneEthInWei))
			balances = append(balances, AddressBalance{
				Address: addr,
				Balance: balanceInEth.String(),
			})
		}
	}

	return balances, nil
}

func weiToEther(wei *big.Int) *big.Float {
	ether := new(big.Float)
	ether.SetString(wei.String())
	return new(big.Float).Quo(ether, big.NewFloat(math.Pow10(18)))
}
