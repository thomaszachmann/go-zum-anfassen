// shorturl ist ein URL-Kürzer mit JSON-API, Metriken und Graceful Shutdown.
// Ab Tag 10 überleben die Links einen Neustart (SQLite).
package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/shorturl/internal/server"
	"example.com/shorturl/internal/store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func main() {
	addr := flag.String("addr", cmp.Or(os.Getenv("SHORTURL_ADDR"), "127.0.0.1:8080"),
		"Listen-Adresse")
	baseURL := flag.String("base-url", "", "öffentliche URL (Default http://ADDR)")
	dbFile := flag.String("db", cmp.Or(os.Getenv("SHORTURL_DB"), "shorturl.db"),
		"SQLite-Datei")
	flag.Parse()
	if *baseURL == "" {
		*baseURL = "http://" + *addr
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(*addr, *baseURL, *dbFile, log); err != nil {
		log.Error("Abbruch", "err", err)
		os.Exit(1)
	}
}

// run öffnet die Datenbank und schließt sie erst nach dem Shutdown.
func run(addr, baseURL, dbFile string, log *slog.Logger) error {
	st, err := store.OpenSQLite(context.Background(), dbFile)
	if err != nil {
		return err
	}
	defer st.Close() //nolint:errcheck // beim Beenden best effort
	log.Info("datenbank bereit", "file", dbFile)
	return serve(addr, baseURL, st, log)
}

func serve(addr, baseURL string, st store.Store, log *slog.Logger) error {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	srv := &http.Server{
		Addr:              addr,
		Handler:           server.New(st, log, baseURL, reg),
		ReadHeaderTimeout: 5 * time.Second, // Schutz gegen Slowloris
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// SIGTERM schickt Kubernetes beim Pod-Stopp, SIGINT ist Strg+C.
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		log.Info("shorturl startet", "addr", addr, "base_url", baseURL)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc: // z. B. Port belegt
		return err
	case <-ctx.Done():
	}
	log.Info("fahre herunter, warte auf laufende Anfragen")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("sauber beendet")
	return nil
}
