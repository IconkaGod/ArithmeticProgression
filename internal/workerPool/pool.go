package pool

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal"
	"github.com/IconkaGod/ArithmeticProgression/internal/queue"
	"github.com/IconkaGod/ArithmeticProgression/internal/service"

	"github.com/shopspring/decimal"
)

type Pool interface {
	StartWorkers()
}

type pool struct {
	queue        queue.Queue
	srv          service.Service
	workersCount int
	ctx          context.Context
	log          *slog.Logger
}

func NewWorkerPool(ctx context.Context, n int, q queue.Queue, srv service.Service, l *slog.Logger) Pool {
	return &pool{
		queue:        q,
		srv:          srv,
		workersCount: n,
		ctx:          ctx,
		log:          l,
	}
}

func (p *pool) StartWorkers() {
	log := p.log.With(
		slog.String("operation", "pool.StartWorkers"),
	)

	for i := 1; i <= p.workersCount; i++ {
		go p.processTask(i)
	}

	log.Info(
		"workers started",
		slog.Int("count_workers", p.workersCount),
	)
}

func (p *pool) processTask(workerId int) {
	log := p.log.With(
		slog.String("operation", "pool.processTask"),
		slog.Int("worker_id", workerId),
	)

	log.Info("worker started")

	defer func() {
		log.Info("worker stopped")
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			taskId, err := p.queue.Pop()
			if err != nil {
				if errors.Is(err, internal.ErrEmptyQueue) {
					log.Debug(
						"queue is empty, retrying",
					)
				} else {
					log.Error(
						"queue pop failed",
						slog.Any("error", err),
					)
					return
				}
				continue
			}

			log = log.With(
				slog.Int64("task_id", taskId),
			)

			task, err := p.srv.GetTaskById(taskId)
			if err != nil {
				if errors.Is(err, internal.ErrNotFound) {
					log.Warn(
						"task ID found in queue but not in cache",
					)
				} else {
					log.Error(
						"failed to fetch task details from service",
						slog.Any("error", err),
					)
				}
				continue
			}

			if task.ElementsCount == 0 {
				log.Info(
					"task processing skipped",
					slog.String("reason", "elements_count is zero"),
				)

				err := p.srv.SetComplete(taskId, time.Now())
				if err != nil {
					log.Error(
						"failed to set task status to COMPLETE",
						slog.Any("error", err),
					)
				}
				continue
			}

			startTime := time.Now()

			err = p.srv.SetInProgress(taskId, startTime)
			if err != nil {
				log.Error(
					"failed to set task status to IN_PROGRESS",
					slog.Any("error", err),
				)
				continue
			}

			log = log.With(
				slog.Time("start_time", startTime),
			)

			log.Info(
				"start of processing",
			)

			task.Result = task.StartNumber

			interval := time.Duration(task.Interval * float64(time.Second))
			ticker := time.NewTicker(interval)

			for i := 1; i < task.ElementsCount; i++ {
				select {
				case <-ticker.C:
					tempRes := decimal.NewFromFloat(task.Result)
					tempDelta := decimal.NewFromFloat(task.Delta)
					tempRes = tempRes.Add(tempDelta)

					task.Result = tempRes.InexactFloat64()

					err = p.srv.UpdateProgress(task.Id, i, task.Result)
					if err != nil {
						log.Warn(
							"failed to update progress task",
							slog.Any("error", err),
						)
					}
				case <-p.ctx.Done():
					ticker.Stop()
					return
				}
			}

			ticker.Stop()

			err = p.srv.SetComplete(taskId, time.Now())
			if err != nil {
				log.Error(
					"failed to set task status to COMPLETE",
					slog.Any("error", err),
				)
				continue
			}

			log.Info(
				"end of processing",
				slog.Time("end_time", time.Now()),
				slog.Float64("duration", time.Since(startTime).Seconds()),
			)
		}
	}
}
