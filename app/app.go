package main

import (
	"context"
	"fmt"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	tracer.Start()

	spanM, ctx := tracer.StartSpanFromContext(ctx, "OrderBookSaver.run")
	defer spanM.Finish()

	wg := sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				span, _ := tracer.StartSpanFromContext(ctx, "dummy")
				select {
				case <-ctx.Done():
					span.Finish()
					return
				case <-time.After(time.Millisecond):
				}
				span.Finish()
			}
		}(ctx, &wg)
	}

	for {
		select {
		case <-ctx.Done():
			goto shutdown
		case <-time.After(time.Second):
		}

		PrintMemUsage()
	}

shutdown:
	wg.Wait()
	fmt.Println("Bye.")
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
// Stolen from https://gist.github.com/j33ty/79e8b736141be19687f565ea4c6f4226
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v", m.NumGC)
	fmt.Printf("\tHeapObjects = %v\n", m.HeapObjects)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
