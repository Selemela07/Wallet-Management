package main

import (
	"bytes"
	"crypto-api/bitcoin"
	"crypto-api/dogecoin"
	"crypto-api/ethereum"
	"crypto-api/helper"
	"crypto-api/litecoin"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Bitcoin  bitcoin.CoinConfig      `json:"bitcoin"`
	Litecoin litecoin.CoinConfig     `json:"litecoin"`
	Dogecoin dogecoin.CoinConfig     `json:"dogecoin"`
	Ethereum ethereum.EthereumConfig `json:"ethereum"`
}

type responseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}
func createHD() {
	mnemonic := helper.GenerateMnemonic()
	//privateKey, publicKey, seed := helper.DeriveKeys(mnemonic, "m/44'/60'/0'/0") // bip32
	privateKey, publicKey, seed := helper.DeriveKeys(mnemonic, "m/44'/60'/0'") // bip44

	fmt.Println("Mnemonic  : ", mnemonic)
	fmt.Println("Private key: ", privateKey)
	fmt.Println("Public key: ", publicKey)
	fmt.Println("BIP39 Seed: ", hex.EncodeToString(seed))
	fmt.Println(helper.GeneratePriv(privateKey, "0/0"))
	fmt.Println(helper.GeneratePub(publicKey, 0))
}
func main() {
	createHD()
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

	//server başladı
	go http.ListenAndServe(":8090", router)

	for {
		bitcoin.RegisterHandlers(router.PathPrefix("/api/bitcoin").Subrouter(), config.Bitcoin)
		litecoin.RegisterHandlers(router.PathPrefix("/api/litecoin").Subrouter(), config.Litecoin)
		dogecoin.RegisterHandlers(router.PathPrefix("/api/dogecoin").Subrouter(), config.Dogecoin)

		ethereum.RegisterHandlers(router.PathPrefix("/api/ethereum").Subrouter(), config.Ethereum)

		bitcoin.InitRedis()
		litecoin.InitRedis()
		dogecoin.InitRedis()

		fig := figure.NewColorFigure("SATOSHITURK.COM", "", "green", true)
		fig.Print()
		color.Cyan("\nWallet Server başarı ile başlatıldı  - Dinlenilen Port :8090")

		color.Red("____________________________")
		color.Red("Wallet Server Sistem Detayları")
		color.Red("____________________________")
		// CPU kullanımı
		cpuPercent, _ := cpu.Percent(0, false)
		color.Green("CPU Kullanımı: %.2f%%\n", cpuPercent[0])

		// RAM kullanımı
		virtualMemory, _ := mem.VirtualMemory()
		color.Yellow("RAM Kullanımı: %.2f%%\n", virtualMemory.UsedPercent)

		// Sistem yükü
		loadAvg, _ := load.Avg()
		color.Magenta("Sistem Yükü (1m, 5m, 15m): %.2f, %.2f, %.2f\n", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)

		// 5 saniye bekle
		time.Sleep(300 * time.Second)

		// Terminali temizle
		fmt.Print("\033[H\033[2J")
	}

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
