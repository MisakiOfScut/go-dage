package executor

import (
	"sync"
)

type Executor interface {
	Execute(func())
	Stop()
}

type DefaultExecutorImpl struct {
	queue  chan func()
	closed bool
	wg     sync.WaitGroup
}

func NewDefaultExecutor(queueLength uint, concurrentLevel uint) Executor {
	if concurrentLevel == 0 {
		concurrentLevel = 1
	}

	d := DefaultExecutorImpl{
		queue:  make(chan func(), queueLength),
		closed: false,
	}
	for i := concurrentLevel; i > 0; i-- {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			for w := range d.queue {
				// TODO(misaki): add a timeout control
				if w != nil {
					w()
				}
			}
		}()
	}
	return &d
}

func (d *DefaultExecutorImpl) Execute(task func()) {
	if !d.closed {
		d.queue <- task
	}
}

// Stop the executor after processing the remaining tasks in the queue. After calling this function,
// you shouldn't call Execute func, which will cause panic.
func (d *DefaultExecutorImpl) Stop() {
	d.closed = true
	close(d.queue)
	d.wg.Wait()
}
