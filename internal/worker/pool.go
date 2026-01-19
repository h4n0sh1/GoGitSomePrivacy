// Package worker provides a worker pool for concurrent processing.
package worker

import (
	"context"
	"sync"
)

// Task represents a unit of work to be processed.
type Task[T any, R any] struct {
	Input  T
	Result R
	Err    error
}

// Pool manages a pool of workers for concurrent task processing.
type Pool[T any, R any] struct {
	workers  int
	taskChan chan *Task[T, R]
	resultCh chan *Task[T, R]
	wg       sync.WaitGroup
	process  func(context.Context, T) (R, error)
}

// NewPool creates a new worker pool.
func NewPool[T any, R any](workers int, process func(context.Context, T) (R, error)) *Pool[T, R] {
	return &Pool[T, R]{
		workers:  workers,
		taskChan: make(chan *Task[T, R], workers*2),
		resultCh: make(chan *Task[T, R], workers*2),
		process:  process,
	}
}

// Start starts the worker pool.
func (p *Pool[T, R]) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}

	// Close result channel when all workers are done
	go func() {
		p.wg.Wait()
		close(p.resultCh)
	}()
}

// worker processes tasks from the task channel.
func (p *Pool[T, R]) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-p.taskChan:
			if !ok {
				return
			}
			result, err := p.process(ctx, task.Input)
			task.Result = result
			task.Err = err

			select {
			case p.resultCh <- task:
			case <-ctx.Done():
				return
			}
		}
	}
}

// Submit submits a task to the pool.
func (p *Pool[T, R]) Submit(input T) {
	p.taskChan <- &Task[T, R]{Input: input}
}

// Close closes the task channel.
func (p *Pool[T, R]) Close() {
	close(p.taskChan)
}

// Results returns the results channel.
func (p *Pool[T, R]) Results() <-chan *Task[T, R] {
	return p.resultCh
}
