package demos

import (
	"fmt"
	"sync"
	"time"
)

func WaitGroupDemo() error {
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(40 * time.Millisecond)
			fmt.Printf("worker %d done\n", id)
		}(i)
	}
	wg.Wait()
	fmt.Println("all workers completed")
	return nil
}
