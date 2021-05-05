package app

import (
	"github.com/w8mr/gomoticasa/config"
	"github.com/w8mr/gomoticasa/server"
	"log"
)

func Start(cfg *config.Config) error {
	if err := server.Run(cfg); err != nil {
		log.Printf("Error in app.Start(): %v", err)
		return err
	}

	return nil
}