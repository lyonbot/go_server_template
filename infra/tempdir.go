package infra

import (
	"log"
	"os"
)

const TEMP_DIR = "./temp"

func init() {
	if err := os.MkdirAll(TEMP_DIR, 0755); err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
}
