package executor

import (
	"github.com/MisakiOfScut/go-dage/internal/utils/log"
	"sync"
)

type Executor interface {
	Execute(func())
	Stop()
}

type DefaultExecutorImpl struct {
	queue chan func()
	wg    sync.WaitGroup
}

func NewDefaultExecutor(queueLength uint, concurrentLevel uint) Executor {
	if queueLength < 32 {
		log.Warn("Executor's queueLength less than 32 maybe lead to bad performance")
	}

	if concurrentLevel == 0 {
		concurrentLevel = 1
	}

	d := DefaultExecutorImpl{
		queue: make(chan func(), queueLength),
	}
	for i := concurrentLevel; i > 0; i-- {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			for w := range d.queue {
				// TODO(misaki): add a timeout control
				w()
			}
		}()
	}
	return &d
}

func (d *DefaultExecutorImpl) Execute(task func()) {
	d.queue <- task
}

// Stop the executor after processing the remaining tasks in the queue. After calling this function,
// you shouldn't call Execute func, which will cause panic.
func (d *DefaultExecutorImpl) Stop() {
	close(d.queue)
	d.wg.Wait()
}
