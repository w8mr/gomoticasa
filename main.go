//go:generate go run gen/gen.go

package main

import (
	"github.com/w8mr/gomoticasa/config"
	"github.com/w8mr/gomoticasa/app"
	"log"
)

func main() {
	cfg := config.New()
	cfg.FromFlags()

	if err := app.Start(cfg); err != nil {
		log.Printf("Error in main(): %v", err)
	}


}
