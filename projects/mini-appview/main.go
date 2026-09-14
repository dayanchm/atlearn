package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dayanchm/mini-appview/internal/consumer"
	appdb "github.com/dayanchm/mini-appview/internal/db"
)

func main() {
	db, err := appdb.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	log.Println("mini appview started")

	if err := consumer.Run(ctx, db); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}

	log.Println("shutdown")
}
