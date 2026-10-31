// Repariert: atomic.Int64 macht das Erhöhen unteilbar.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var wg sync.WaitGroup
	var ok atomic.Int64
	for range 1000 {
		wg.Go(func() { ok.Add(1) })
	}
	wg.Wait()
	fmt.Println("ok =", ok.Load())
}
