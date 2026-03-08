// examples/go-project/main.go
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env if it exists
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "sqlite:memory:" // default for example
	}

	fmt.Printf("Rocket Project Starting...\n")
	fmt.Printf("Connected to: %s\n", dbURL)
	fmt.Println("Ready to launch!")
}
