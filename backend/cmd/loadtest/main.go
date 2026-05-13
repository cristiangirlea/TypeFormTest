package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080/form/1d7b3681", "URL to test")
	concurrency := flag.Int("c", 100, "Number of concurrent workers")
	duration := flag.Duration("d", 10*time.Second, "Duration of the test")
	flag.Parse()

	fmt.Printf("Starting load test on %s\n", *url)
	fmt.Printf("Concurrency: %d, Duration: %s\n", *concurrency, *duration)

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        *concurrency,
			MaxIdleConnsPerHost: *concurrency,
			MaxConnsPerHost:     *concurrency,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	var successCount int64
	var errorCount int64
	var totalLatency int64 // in nanoseconds

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(*concurrency)

	for i := range *concurrency {
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					reqStart := time.Now()
					req, _ := http.NewRequestWithContext(ctx, "GET", *url, nil)
					resp, err := client.Do(req)

					if err != nil {
						atomic.AddInt64(&errorCount, 1)
						// Backoff slightly on error to prevent spinning and thread exhaustion
						time.Sleep(10 * time.Millisecond)
						continue
					}

					// Drain and close body
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						atomic.AddInt64(&successCount, 1)
						atomic.AddInt64(&totalLatency, int64(time.Since(reqStart)))
					} else {
						atomic.AddInt64(&errorCount, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)

	success := atomic.LoadInt64(&successCount)
	errors := atomic.LoadInt64(&errorCount)
	totalReqs := success + errors
	rps := float64(success) / totalTime.Seconds()

	avgLatency := time.Duration(0)
	if success > 0 {
		avgLatency = time.Duration(atomic.LoadInt64(&totalLatency) / success)
	}

	fmt.Println("\n--- Results ---")
	fmt.Printf("Total Requests: %d\n", totalReqs)
	fmt.Printf("Successful:     %d\n", success)
	fmt.Printf("Errors:         %d\n", errors)
	fmt.Printf("RPS (Success):  %.2f req/s\n", rps)
	fmt.Printf("Avg Latency:    %s\n", avgLatency)
	fmt.Printf("Total Time:     %s\n", totalTime)
}
