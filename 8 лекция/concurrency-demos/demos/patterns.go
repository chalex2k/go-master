package demos

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func FanOutDemo() error {
	jobs := make(chan int)
	results := make(chan int)

	for w := 1; w <= 3; w++ {
		go func() {
			for j := range jobs {
				results <- j * j
				fmt.Printf("worker %d processed %d\n", w, j)
			}
		}()
	}

	go func() {
		for i := 1; i <= 6; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	for i := 0; i < 6; i++ {
		fmt.Printf("result %d\n", <-results)
	}
	return nil
}

func FanInDemo() error {
	left := make(chan int)
	right := make(chan int)

	go func() {
		defer close(left)
		for i := 1; i <= 3; i++ {
			left <- i
		}
	}()
	go func() {
		defer close(right)
		for i := 10; i <= 12; i++ {
			right <- i
		}
	}()

	out := merge(left, right)
	for v := range out {
		fmt.Printf("merged: %d\n", v)
	}
	return nil
}

func PipelineFanOutFanInDemo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	jobs := make(chan int)
	workerOut := make([]<-chan int, 0, 3)

	for i := 0; i < 3; i++ {
		workerOut = append(workerOut, worker(ctx, i+1, jobs))
	}

	go func() {
		defer close(jobs)
		for i := 1; i <= 9; i++ {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	out := merge(workerOut...)
	sum := 0
	for v := range out {
		sum += v
	}
	fmt.Printf("pipeline sum=%d\n", sum)
	return nil
}

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	wg.Add(len(cs))
	for _, c := range cs {
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func worker(ctx context.Context, id int, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case j, ok := <-in:
				if !ok {
					return
				}
				result := j * j
				fmt.Printf("pipeline worker %d handled %d\n", id, j)
				select {
				case out <- result:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
