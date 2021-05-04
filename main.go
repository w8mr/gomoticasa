//go:generate go run gen/gen.go

package main

import (
	"github.com/w8mr/gomoticasa/config"
	"github.com/w8mr/gomoticasa/server"
	"log"
)

func main() {
	cfg := config.New()
	cfg.FromFlags()

	if err := server.Run(cfg); err != nil {
		log.Printf("Error in main(): %v", err)
	}
}
