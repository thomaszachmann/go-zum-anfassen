// genlog erzeugt ein reproduzierbares Access-Log im Combined-Format.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"time"
)

func main() {
	n := flag.Int("n", 1_000_000, "Anzahl Zeilen")
	flag.Parse()
	rng := rand.New(rand.NewPCG(42, 2026)) // fester Seed: überall dieselbe Datei
	paths := []string{"/", "/login", "/api/users", "/api/orders", "/api/cart",
		"/healthz", "/metrics", "/static/app.js", "/api/search", "/admin"}
	methods := []string{"GET", "GET", "GET", "POST", "PUT", "DELETE"}
	codes := []int{200, 200, 200, 200, 201, 204, 301, 404, 500, 503}
	t := time.Date(2026, 10, 10, 8, 0, 0, 0, time.FixedZone("", 2*3600))
	const layout = "02/Jan/2006:15:04:05 -0700"
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for i := range *n {
		fmt.Fprintf(w, "10.0.%d.%d - - [%s] \"%s %s HTTP/1.1\" %d %d \"-\" \"k6\"\n",
			rng.IntN(4), rng.IntN(255),
			t.Add(time.Duration(i)*time.Millisecond).Format(layout),
			methods[rng.IntN(len(methods))], paths[rng.IntN(len(paths))],
			codes[rng.IntN(len(codes))], rng.IntN(50_000))
	}
}
