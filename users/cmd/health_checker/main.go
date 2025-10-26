package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nimbo1999/financeiro/users/internal/handler"
)

func main() {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(timeoutCtx, "GET", "http://localhost/health", nil)
	if err != nil {
		log.Fatalln(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	defer response.Body.Close()
	var healthResponse handler.HealthResponse
	if err := json.NewDecoder(response.Body).Decode(&healthResponse); err != nil {
		log.Fatalln(err)
	}
	if healthResponse.Status != "healthy" {
		log.Fatalf("Service is not healthy: %s\n", healthResponse.Status)
	}
	log.Println("Service users is healthy")
	os.Exit(0)
}
