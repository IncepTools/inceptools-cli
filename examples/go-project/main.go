// examples/go-project/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	// Simple example: relies on environment variables set in the shell or .env
	// In a real project, you might use a library like godotenv here.

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "sqlite:memory:" // default for example
	}

	fmt.Printf("Rocket Project Starting...\n")
	fmt.Printf("Connected to: %s\n", dbURL)
	fmt.Println("Ready to launch!")
}
