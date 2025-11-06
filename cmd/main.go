package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/IconkaGod/ArithmeticProgression/internal/api/handlers"
	"github.com/IconkaGod/ArithmeticProgression/internal/queue"
	"github.com/IconkaGod/ArithmeticProgression/internal/repository"
	"github.com/IconkaGod/ArithmeticProgression/internal/router"
	"github.com/IconkaGod/ArithmeticProgression/internal/service"
	workers "github.com/IconkaGod/ArithmeticProgression/internal/workerPool"
)

func main() {
	var workersCount int
	flag.IntVar(&workersCount, "workers", 1, "count of workers")
	flag.Parse()

	if workersCount <= 0 {
		log.Fatal("The number of workers cannot be <= 0")
	}

	ctx, cancel := context.WithCancel(context.Background())

	repo := repository.NewRepository(ctx)

	q := queue.NewQueue()

	srv := service.NewService(q, repo)

	hand := handlers.NewHandlers(srv)

	pool := workers.NewWorkerPool(ctx, workersCount, q, srv)
	pool.StartWorkers()

	r := router.NewRouter(hand)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Println("Starting server")

	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatalf("failed to start server:%v", err)
		}
	}()

	<-stop

	log.Println("Shutting down service...")
	cancel()
}
