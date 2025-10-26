package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

var ErrNotPositiveWorkers = errors.New("number of workers must be positive")

type Task func() error

func worker(jobs <-chan Task, wg *sync.WaitGroup, errLimit int32, errorCount *atomic.Int32) {
	defer wg.Done()
	for job := range jobs {
		if errorCount.Load() >= errLimit {
			return
		}
		if err := job(); err != nil {
			errorCount.Add(1)
		}
	}
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrNotPositiveWorkers
	}
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	jobs := make(chan Task)
	var wg sync.WaitGroup
	var errorCount atomic.Int32
	errLimit := int32(m)

	for i := 1; i <= n; i++ {
		wg.Add(1)
		go worker(jobs, &wg, errLimit, &errorCount)
	}

	for i, task := range tasks {
		if errorCount.Load() >= errLimit {
			break
		}
		fmt.Printf("Задача %d отправлена в канал\n", i)
		jobs <- task
	}

	close(jobs)
	wg.Wait()

	fmt.Println("Все задачи завершены")

	if errorCount.Load() >= errLimit {
		return ErrErrorsLimitExceeded
	}
	return nil
}
