package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

func main() {
	inputCh := make(chan int)
	workCh := make(chan int)
	outputCh := make(chan []int)
	done := make(chan bool)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go scheduler(ctx, inputCh, workCh)
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(ctx, workCh, outputCh)
	}

	wg.Add(1)
	inputCh <- 10

	go func() {
		wg.Wait()
		done <- true
	}()

	seen := make(map[int]bool)

	for {
		select {
		case <-done:
			return
		case nums := <-outputCh:
			for _, n := range nums {
				if _, exists := seen[n]; !exists && n%2 == 0 {
					wg.Add(1)
					inputCh <- n
				}
				fmt.Println(n)
				seen[n] = true
			}
			wg.Done()
		}
	}
}

func scheduler(ctx context.Context, input chan int, work chan int) {
	var queue []int
	for {
		if len(queue) == 0 {
			select {
			case <-ctx.Done():
				return
			case i := <-input:
				queue = append(queue, i)
			}
		} else {
			select {
			case <-ctx.Done():
				return
			case i := <-input:
				queue = append(queue, i)
			case work <- queue[0]:
				queue = queue[1:]
			}
		}
	}
}

func worker(ctx context.Context, work chan int, output chan []int) {
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-work:
			var nums []int
			for i := 1; i <= n; i++ {
				nums = append(nums, i)
			}
			output <- nums
		}
	}
}
