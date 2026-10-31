package check

import (
	"context"
	"sync"
	"time"
)

// Run prüft alle URLs mit c.Workers Goroutines. Ergebnisse in Eingabe-Reihenfolge.
func (c *Checker) Run(ctx context.Context, urls []string) []Result {
	results := make([]Result, len(urls))
	jobs := make(chan int) // Index in urls

	// Rate-Limit: jeder Worker holt sich vor der Anfrage einen Tick.
	var tick <-chan time.Time
	if c.RPS > 0 {
		t := time.NewTicker(time.Duration(float64(time.Second) / c.RPS))
		defer t.Stop()
		tick = t.C
	}

	var wg sync.WaitGroup
	for range max(c.Workers, 1) {
		wg.Go(func() {
			for i := range jobs { // endet, wenn jobs geschlossen ist
				if tick != nil {
					select {
					case <-tick:
					case <-ctx.Done():
						results[i] = Result{URL: urls[i], Err: ctx.Err().Error()}
						continue
					}
				}
				results[i] = c.One(ctx, urls[i]) // eigener Index: kein Lock nötig
			}
		})
	}

	for i := range urls {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}
