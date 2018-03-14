//go:generate go run gen/gen.go

package main

import (
	"w8mr.nl/go_my_home/config"
	"w8mr.nl/go_my_home/server"
	"log"
)

func main() {
	cfg := config.New()
	cfg.FromFlags()

	if err := server.Run(cfg); err != nil {
		log.Printf("Error in main(): %v", err)
	}
}
