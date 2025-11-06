package workers

import (
	"context"
	"time"

	"github.com/IconkaGod/ArithmeticProgression/internal/queue"
	"github.com/IconkaGod/ArithmeticProgression/internal/service"
)

type Pool interface {
	StartWorkers()
}

type pool struct {
	queue        queue.Queue
	srv          service.Service
	workersCount int
	ctx          context.Context
}

func NewWorkerPool(ctx context.Context, n int, q queue.Queue, srv service.Service) Pool {
	return &pool{
		queue:        q,
		srv:          srv,
		workersCount: n,
		ctx:          ctx,
	}
}

func (p *pool) StartWorkers() {
	for i := 0; i < p.workersCount; i++ {
		go p.processTask()
	}
}

func (p *pool) processTask() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			taskId, err := p.queue.Pop()
			if err != nil {
				time.Sleep(300 * time.Millisecond)
				continue
			}

			task, err := p.srv.GetTaskById(taskId)
			if err != nil {
				continue
			}

			if task.ElementsCount == 0 {
				p.srv.SetComplete(taskId, time.Now())
				continue
			}

			p.srv.SetInProgress(taskId, time.Now())

			interval := time.Duration(task.Interval * float64(time.Second))
			ticker := time.NewTicker(interval)

			task.Result = task.StartNumber

			for i := 0; i < task.ElementsCount; i++ {
				select {
				case <-ticker.C:
					task.Result += task.Delta
					task.Iteration++

					p.srv.UpdateProgress(task.Id, task.Iteration, task.Result)
				case <-p.ctx.Done():
					ticker.Stop()
					return
				}
			}

			ticker.Stop()

			err = p.srv.SetComplete(taskId, time.Now())
			if err != nil {
				continue
			}

		}
	}
}
