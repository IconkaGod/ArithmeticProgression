package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal/api/handlers"
	"github.com/IconkaGod/ArithmeticProgression/internal/cache"
	"github.com/IconkaGod/ArithmeticProgression/internal/queue"
	"github.com/IconkaGod/ArithmeticProgression/internal/router"
	"github.com/IconkaGod/ArithmeticProgression/internal/service"
	pool "github.com/IconkaGod/ArithmeticProgression/internal/workerPool"
)

func main() {
	var workersCount int
	flag.IntVar(&workersCount, "workers", 1, "count of workers")
	flag.Parse()

	if workersCount <= 0 {
		log.Fatal("The number of workers cannot be <= 0")
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cache := cache.NewCache(ctx, logger, 2*time.Second)

	q := queue.NewQueue(logger)

	srv := service.NewService(q, cache, logger)

	hand := handlers.NewHandlers(srv, logger)

	pool := pool.NewWorkerPool(ctx, workersCount, q, srv, logger)
	pool.StartWorkers()

	r := router.NewRouter(hand)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting server")

	go func() {
		server := &http.Server{
			Addr:         ":8080",
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 20 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("failed to start server:%v", err)
		}
	}()

	<-sigs

	log.Println("Shutting down service...")
	cancel()
}
