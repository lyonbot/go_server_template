package main

import (
	"log"
	"os"

	"lyonbot.github.com/my_app/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := server.NewServer()
	if err := server.Start("0.0.0.0:" + port); err != nil {
		log.Fatal(err)
	}
}
