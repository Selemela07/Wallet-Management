package ethereum

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"log"
	"math/big"
	"net/http"
	"strconv"
)

// Redis client
var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379", // replace with your Redis address and port
	Password: "",               // replace with your password if any
	DB:       0,                // use default DB
})

func RegisterHandlers(router *mux.Router, cfg EthereumConfig) {
	router.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		getBlock(w, r, cfg)
	}).Methods("POST")

	router.HandleFunc("/blockrange", func(w http.ResponseWriter, r *http.Request) {
		getBlockRange(w, r, cfg)
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

	err = getAllTransactionAddresses(client, start, end)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching addresses: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "All addresses processed and set in Redis"})
}

func getAllTransactionAddresses(client *ethclient.Client, startBlock, endBlock *big.Int) error {
	for blockNumber := new(big.Int).Set(startBlock); blockNumber.Cmp(endBlock) <= 0; blockNumber.Add(blockNumber, big.NewInt(1)) {
		// Print the current block number in the same line
		fmt.Printf("\rProcessing block: %s", blockNumber.String())

		block, err := client.BlockByNumber(ctx, blockNumber)
		if err != nil {
			log.Printf("\nError fetching block %s: %v", blockNumber, err)
			return err
		}
		for _, tx := range block.Transactions() {
			to := tx.To()
			if to != nil {
				balance, err := client.BalanceAt(ctx, *to, nil)
				if err != nil {
					log.Printf("\nError fetching balance for address %s: %v", to.Hex(), err)
					continue
				}

				balanceInEth := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(big.NewInt(1e18)))
				err = rdb.Set(ctx, to.Hex(), balanceInEth.String(), 0).Err()
				if err != nil {
					log.Printf("\nError setting balance in Redis for address %s: %v", to.Hex(), err)
				}
			}
		}
	}
	return nil
}
