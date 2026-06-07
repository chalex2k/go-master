package demos

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func ErrGroupDemo() error {
	type task struct {
		name  string
		delay time.Duration
		fail  bool
	}
	tasks := []task{
		{"task-A", 50 * time.Millisecond, false},
		{"task-B", 80 * time.Millisecond, true},
		{"task-C", 150 * time.Millisecond, false},
	}

	g, ctx := errgroup.WithContext(context.Background())
	for _, t := range tasks {
		g.Go(func() error {
			select {
			case <-time.After(t.delay):
				if t.fail {
					return fmt.Errorf("%s failed", t.name)
				}
				fmt.Printf("%s done\n", t.name)
				return nil
			case <-ctx.Done():
				fmt.Printf("%s canceled\n", t.name)
				return ctx.Err()
			}
		})
	}

	err := g.Wait()
	fmt.Printf("errgroup wait returned: %v\n", err)
	return nil
}
