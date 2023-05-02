package bitcoin

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
	router.HandleFunc("/getbalance", getBalanceHandler(config)).Methods("POST")
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
			Address               string  `json:"address"`
			Amount                float64 `json:"amount"`
			Comment               string  `json:"comment,omitempty"`
			CommentTo             string  `json:"comment_to,omitempty"`
			SubtractFeeFromAmount bool    `json:"subtract_fee_from_amount,omitempty"`
			Replaceable           *bool   `json:"replaceable,omitempty"`
			ConfTarget            *int    `json:"conf_target,omitempty"`
			EstimateMode          string  `json:"estimate_mode,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&sendReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		params := []interface{}{
			sendReq.Address,
			sendReq.Amount,
		}
		if sendReq.Comment != "" {
			params = append(params, sendReq.Comment)
		}
		if sendReq.CommentTo != "" {
			params = append(params, sendReq.CommentTo)
		}
		params = append(params, sendReq.SubtractFeeFromAmount)
		if sendReq.Replaceable != nil {
			params = append(params, *sendReq.Replaceable)
		}
		if sendReq.ConfTarget != nil {
			params = append(params, *sendReq.ConfTarget)
		}
		if sendReq.EstimateMode != "" {
			params = append(params, sendReq.EstimateMode)
		}

		response, err := makeJSONRPCRequest(config, "sendtoaddress", params)
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
