package main

import (
	"bytes"
	"crypto-api/bitcoin"
	"crypto-api/dogecoin"
	"crypto-api/litecoin"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Config struct {
	Bitcoin  bitcoin.CoinConfig  `json:"bitcoin"`
	Litecoin litecoin.CoinConfig `json:"litecoin"`
	Dogecoin dogecoin.CoinConfig `json:"dogecoin"`
}

type CoinConfig struct {
	RPCIP       string `json:"rpcip"`
	RPCPort     string `json:"rpcport"`
	RPCUser     string `json:"rpcuser"`
	RPCPassword string `json:"rpcpassword"`
}

type responseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer configFile.Close()

	var config Config
	json.NewDecoder(configFile).Decode(&config)

	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	bitcoin.RegisterHandlers(router.PathPrefix("/api/bitcoin").Subrouter(), config.Bitcoin)
	litecoin.RegisterHandlers(router.PathPrefix("/api/litecoin").Subrouter(), config.Litecoin)
	dogecoin.RegisterHandlers(router.PathPrefix("/api/dogecoin").Subrouter(), config.Dogecoin)

	http.ListenAndServe(":8080", router)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rw := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}
		next.ServeHTTP(rw, r)
		duration := time.Since(startTime)

		clientIP := color.New(color.FgGreen).SprintFunc()
		method := color.New(color.FgYellow).SprintFunc()
		path := color.New(color.FgCyan).SprintFunc()
		durationColor := color.New(color.FgMagenta).SprintFunc()
		response := color.New(color.FgWhite).SprintFunc()

		log.Printf("%s %s %s %s\n%s", clientIP(r.RemoteAddr), method(r.Method), path(r.RequestURI), durationColor(duration), response(rw.body.String()))
	})
}
