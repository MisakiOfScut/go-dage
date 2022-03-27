package executor

import (
	"fmt"
	"sync"
	"testing"
)

func TestCreateExecutor(t *testing.T) {
	e := NewDefaultExecutor(0, 0)
	defer e.Stop()
}

func TestExecuteFunction(t *testing.T) {
	e := NewDefaultExecutor(32, 8)
	e.Execute(func() {
		fmt.Println("test")
	})
	e.Stop()
	e.Execute(func() {
		fmt.Println("this log should not print")
	})
}

func TestConcurrentExecute(t *testing.T) {
	e := NewDefaultExecutor(32, 8)
	defer e.Stop()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			e.Execute(func() {
				fmt.Printf("number %d\n", j)
			})
		}(i)
	}
	wg.Wait()
}
