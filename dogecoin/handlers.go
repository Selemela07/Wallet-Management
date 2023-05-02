package dogecoin

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type CoinConfig struct {
	RPCIP       string `json:"rpcip"`
	RPCPort     string `json:"rpcport"`
	RPCUser     string `json:"rpcuser"`
	RPCPassword string `json:"rpcpassword"`
}

type JSONRPCRequest struct {
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	JsonRpc string      `json:"jsonrpc"`
}

func RegisterHandlers(router *mux.Router, config CoinConfig) {
	router.HandleFunc("/tx/{txid}", transactionHandler(config)).Methods("GET")
	router.HandleFunc("/getbalance", getBalanceHandler(config)).Methods("GET")
	router.HandleFunc("/sendtransactions", sendTransactionsHandler(config)).Methods("POST")
}

func transactionHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txID := mux.Vars(r)["txid"]
		response, err := makeJSONRPCRequest(config, "getrawtransaction", []interface{}{txID, 1})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func getBalanceHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := makeJSONRPCRequest(config, "getbalance", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func sendTransactionsHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sendReq struct {
			Amounts map[string]float64 `json:"amounts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&sendReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response, err := makeJSONRPCRequest(config, "sendmany", []interface{}{"", sendReq.Amounts})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func makeJSONRPCRequest(config CoinConfig, method string, params interface{}) ([]byte, error) {
	request := JSONRPCRequest{
		ID:      1,
		JsonRpc: "2.0",
		Method:  method,
		Params:  params,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := "http://" + config.RPCUser + ":" + config.RPCPassword + "@" + config.RPCIP + ":" + config.RPCPort
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}
