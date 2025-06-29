package producer

import (
	"context"
	"log/slog"
	"mist/src/faults"
	"sync"
	"sync/atomic"
)

type WorkerPool struct {
	workers  int
	jobQueue chan Job
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	closed   int32
}

func NewWorkerPool(w int, qSize int) *WorkerPool {
	ctx := context.Background()

	wp := &WorkerPool{
		workers:  w,
		jobQueue: make(chan Job, qSize),
		ctx:      ctx,
	}

	return wp
}

func (wp *WorkerPool) StartWorkers() {
	// Start the worker pool by initializing the context and starting workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.jobHandler(i)
	}
}

func (wp *WorkerPool) Stop() {
	atomic.StoreInt32(&wp.closed, 1) // Mark the worker pool as closed

	for range wp.workers {
		wp.AddJob(NewStopWorkerJob(wp.ctx)) // Add stop worker jobs to unblock workers waiting on the job queue
	}

	wp.wg.Wait()       // Wait for all workers to finish
	close(wp.jobQueue) // Close the job queue channel
}

func (wp *WorkerPool) jobHandler(worker int) {
	for job := range wp.jobQueue {
		if _, ok := job.(*StopWorkerJob); ok {
			// If the job is a stop worker job, we exit the loop
			wp.wg.Done()
			return
		}

		if err := job.Execute(worker); err != nil {
			faults.LogError(job.Ctx(), faults.ExtendError(err))
		}
	}
}

func (wp *WorkerPool) AddJob(job Job) {

	if _, ok := job.(*StopWorkerJob); ok {
		// if job is a stop worker job, we just add it to the queue
		wp.jobQueue <- job
		return
	}

	if atomic.LoadInt32(&wp.closed) == 1 {
		faults.UnknownError("WorkerPool is closed", slog.LevelWarn)
		return
	}

	wp.jobQueue <- job
	// if the channel is full, we wait until space is available
}

func (wp *WorkerPool) GetJobQueueSize() int {
	return len(wp.jobQueue)
}
