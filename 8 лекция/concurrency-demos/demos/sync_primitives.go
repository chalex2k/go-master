package demos

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func MutexDemo() error {
	counter := map[string]int{"hits": 0}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter["hits"]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("mutex counter hits=%d\n", counter["hits"])
	return nil
}

func RWMutexDemo() error {
	type safeStore struct {
		mu sync.RWMutex
		m  map[string]int
	}
	store := safeStore{m: map[string]int{"a": 1}}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			store.mu.RLock()
			v := store.m["a"]
			store.mu.RUnlock()
			fmt.Printf("reader %d saw %d\n", id, v)
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		store.mu.Lock()
		store.m["a"] = 2
		store.mu.Unlock()
		fmt.Println("writer updated value to 2")
	}()

	wg.Wait()
	return nil
}

func AtomicDemo() error {
	var total int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&total, 1)
		}()
	}
	wg.Wait()
	fmt.Printf("atomic total=%d\n", total)
	return nil
}
