package demos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func ChannelsProducerConsumerDemo() error {
	jobs := make(chan int, 4)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := range jobs {
			fmt.Printf("processed job %d\n", j)
		}
	}()

	for i := 1; i <= 5; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return nil
}

func SelectWithContextDemo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	dataCh := make(chan string, 1)
	go func() {
		time.Sleep(60 * time.Millisecond)
		dataCh <- "ready"
	}()

	select {
	case msg := <-dataCh:
		fmt.Printf("select got message: %s\n", msg)
		return nil
	case <-time.After(200 * time.Millisecond):
		return errors.New("timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
}
