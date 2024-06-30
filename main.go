package main

import (
	"fmt"
	"url-tester/api"
)

func main() {
	router := api.NewRouter()

	fmt.Printf("Starting server on port %d", 4044)
	if err := router.Run(":4044"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
