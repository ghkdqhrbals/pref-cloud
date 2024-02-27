package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	dbConfig     DatabaseConfig
	serverConfig ServerConfig
)

func init() {
	serverConfig = ServerConfig{
		Port: 8080,
	}
	dbConfig = DatabaseConfig{
		Host:     "localhost",
		User:     "test",
		Password: "test",
		DBName:   "test",
		Port:     5432,
	}
}

func main() {
	ctx, cancle := context.WithCancel(context.Background())

	db := dbConfig.connectDB()

	go startServer(ctx, cancle, &serverConfig, db)

	// Wait for interrupt signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	cancle()
	fmt.Println("Received interrupt signal. Shutting down...")
}

func (c *DatabaseConfig) connectDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Seoul", c.Host, c.User, c.Password, c.DBName, c.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	if err := db.AutoMigrate(&ResultsJSON{}); err != nil {
		panic("failed to migrate database")
	}
	return db
}

func startServer(ctx context.Context, cancel context.CancelFunc, config *ServerConfig, db *gorm.DB) {
	// Setup routes
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req TestDto
			err := json.NewDecoder(r.Body).Decode(&req)

			if err != nil {
				http.Error(w, "Failed to decode request body", http.StatusBadRequest)
				return
			}
			// Run test method
			test(ctx, *db, req.NumUsers, req.NumReqs, req.URL, cancel)
			w.Write([]byte("Success!"))
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	addr := fmt.Sprintf(":%d", config.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	cancel()
}

func test(ctx context.Context, db gorm.DB, numUsers, numRequestsPerUser int, url string, cancel context.CancelFunc) {

	results := make(chan RequestMetrics, numUsers*numRequestsPerUser) // Buffer the channel to prevent blocking
	client := createHttpClient()

	var wg sync.WaitGroup
	fmt.Printf("startTest")
	startTest(ctx, numUsers, numRequestsPerUser, url, &wg, results, client)

	wg.Wait() // WaitGroup이 모든 고루틴이 종료될 때까지 대기
	testResults := ProcessResults(results, numUsers)
	testResults.Method = "POST"
	testResults.URL = url

	marshalJSON := testResults.MarshalJSON()

	if err := db.Create(&marshalJSON).Error; err != nil {
		panic("failed to create data")
	}

}

func createHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 1,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
