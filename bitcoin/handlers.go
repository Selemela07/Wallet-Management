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
	router.HandleFunc("/tx", transactionHandler(config)).Methods("POST")
	router.HandleFunc("/getnewaddress", getGetNewAddressHandler(config)).Methods("POST")
	router.HandleFunc("/getbalance", getBalanceHandler(config)).Methods("POST")
	router.HandleFunc("/getaddressbalance", getAddressBalanceHandler(config)).Methods("POST")
	router.HandleFunc("/sendtransactions", sendTransactionsHandler(config)).Methods("POST")
	router.HandleFunc("/getreceivedbyaddress", getAddressReceivedHandler(config)).Methods("POST")
	router.HandleFunc("/listunspent", listUnspentHandler(config)).Methods("POST")
	router.HandleFunc("/listtransactionsbyaddress", listTransactionsByAddressHandler(config)).Methods("POST")

}

func transactionHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TXID string `json:"txid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.TXID == "" {
			http.Error(w, "TXID parameter is required", http.StatusBadRequest)
			return
		}

		response, err := makeJSONRPCRequest(config, "gettransaction", []interface{}{req.TXID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func getGetNewAddressHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			JsonRpc string        `json:"jsonrpc"`
			ID      string        `json:"id"`
			Method  string        `json:"method"`
			Params  []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.Params) != 2 {
			http.Error(w, "Two parameters (label and address_type) are required", http.StatusBadRequest)
			return
		}

		label, ok := req.Params[0].(string)
		if !ok {
			http.Error(w, "The first parameter must be a string (label)", http.StatusBadRequest)
			return
		}

		addressType, ok := req.Params[1].(string)
		if !ok {
			http.Error(w, "The second parameter must be a string (address_type)", http.StatusBadRequest)
			return
		}

		response, err := makeJSONRPCRequest(config, "getnewaddress", []interface{}{label, addressType})
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
		var req struct {
			JsonRpc string        `json:"jsonrpc"`
			ID      string        `json:"id"`
			Method  string        `json:"method"`
			Params  []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.Params) != 1 {
			http.Error(w, "One parameter (address) is required", http.StatusBadRequest)
			return
		}

		address, ok := req.Params[0].(string)
		if !ok {
			http.Error(w, "The parameter must be a string (address)", http.StatusBadRequest)
			return
		}

		response, err := makeJSONRPCRequest(config, "getbalance", []interface{}{address})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func getAddressBalanceHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			JsonRpc string        `json:"jsonrpc"`
			ID      int64         `json:"id"`
			Method  string        `json:"method"`
			Params  []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.Params) != 1 {
			http.Error(w, "One parameter (address) is required", http.StatusBadRequest)
			return
		}

		address, ok := req.Params[0].(string)
		if !ok {
			http.Error(w, "The parameter must be a string (address)", http.StatusBadRequest)
			return
		}

		response, err := makeJSONRPCRequest(config, "listunspent", []interface{}{0, 9999999, []string{address}})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var listUnspentResponse struct {
			Result []struct {
				Amount float64 `json:"amount"`
			} `json:"result"`
			Error json.RawMessage `json:"error"`
			ID    int64           `json:"id"`
		}
		if err := json.Unmarshal(response, &listUnspentResponse); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var balance float64
		for _, unspent := range listUnspentResponse.Result {
			balance += unspent.Amount
		}

		balanceResponse := struct {
			Result float64     `json:"result"`
			Error  interface{} `json:"error"`
			ID     int64       `json:"id"`
		}{
			Result: balance,
			Error:  nil,
			ID:     req.ID,
		}

		responseBytes, err := json.Marshal(balanceResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseBytes)
	}
}

func getAddressReceivedHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Address string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Address == "" {
			http.Error(w, "Address parameter is required", http.StatusBadRequest)
			return
		}

		response, err := makeJSONRPCRequest(config, "getreceivedbyaddress", []interface{}{req.Address})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func listUnspentHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			MinConf   *int     `json:"minconf,omitempty"`
			MaxConf   *int     `json:"maxconf,omitempty"`
			Addresses []string `json:"addresses,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		params := []interface{}{}
		if req.MinConf != nil {
			params = append(params, *req.MinConf)
		}
		if req.MaxConf != nil {
			params = append(params, *req.MaxConf)
		}
		if len(req.Addresses) > 0 {
			params = append(params, req.Addresses)
		}

		response, err := makeJSONRPCRequest(config, "listunspent", params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func listTransactionsByAddressHandler(config CoinConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Address string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Address == "" {
			http.Error(w, "Address is required", http.StatusBadRequest)
			return
		}

		params := []interface{}{req.Address}
		response, err := makeJSONRPCRequest(config, "listtransactions", params)
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
		JsonRpc: "1.0", // JSON-RPC versiyonunu 1.0 olarak güncelledik
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
