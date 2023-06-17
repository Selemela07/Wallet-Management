package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4"
)

type RedisKeyValue struct {
	Key   string
	Value string
}

func main() {
	// Redis bağlantı ayarları
	redisClient := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	})

	// PostgreSQL bağlantı ayarları
	connConfig, err := pgx.ParseConfig("postgresql://postgres:asdasdasd@localhost:5432/bsc") // PostgreSQL bağlantı bilgileri
	if err != nil {
		panic(err)
	}

	// Redis'teki tüm anahtarları al
	keys, err := redisClient.Keys("*").Result()
	if err != nil {
		panic(err)
	}

	// Toplu veri işleme için bir slice oluştur
	var redisData []RedisKeyValue

	// Her 500.000 veride bir toplu işlem yap
	const batchSize = 30000
	var wg sync.WaitGroup

	for i, key := range keys {
		value, err := redisClient.Get(key).Result()
		if err != nil {
			fmt.Printf("Hata: %v\n", err)
			continue
		}

		redisData = append(redisData, RedisKeyValue{Key: key, Value: value})

		if (i+1)%batchSize == 0 || i == len(keys)-1 {
			wg.Add(1)
			go func(data []RedisKeyValue) {
				defer wg.Done()
				err := insertBatchIntoPostgreSQL(context.Background(), connConfig, data)
				if err != nil {
					fmt.Printf("Hata: %v\n", err)
				}
			}(redisData)

			redisData = nil // Verileri sıfırla
		}
	}

	wg.Wait()

	fmt.Println("Veriler PostgreSQL'e aktarıldı.")
}

// PostgreSQL'e veriyi toplu olarak eklemek için fonksiyon
func insertBatchIntoPostgreSQL(ctx context.Context, config *pgx.ConnConfig, data []RedisKeyValue) error {
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// Toplu INSERT işlemi için bir dize oluştur
	query := "INSERT INTO address (address, balance) VALUES "
	values := make([]interface{}, 0, len(data)*2)

	for i, kv := range data {
		query += fmt.Sprintf("($%d, $%d),", i*2+1, i*2+2)
		values = append(values, kv.Key, kv.Value)
	}

	query = query[:len(query)-1] // Son virgülü kaldır

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}
