package daemon

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"../api"
	"../db"
	"../model"
)

type Config struct {
	ListenHost string

	Db  db.Config
	API api.Config
}

func Run(cfg *Config) error {
	log.Printf("Starting, HTTP on: %s\n", cfg.ListenHost)

	db, err := db.InitDb(cfg.Db)
	if err != nil {
		log.Printf("Error initializing database: %v\n", err)
		return err
	}

	m := model.New(db)

	l, err := net.Listen("tcp", cfg.ListenHost)
	if err != nil {
		log.Printf("Error creating listener: %v\n", err)
		return err
	}

	api.Start(cfg.API, m, l)

	waitForSignal()

	return nil
}

func waitForSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	log.Printf("Got signal: %v, exiting.", s)
}
