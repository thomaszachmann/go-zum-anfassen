package server

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// statusRecorder merkt sich den Statuscode, den der Handler schreibt.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logRequests schreibt pro Anfrage eine strukturierte Logzeile.
func logRequests(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.InfoContext(r.Context(), "request", "method", r.Method,
			"path", r.URL.Path, "status", rec.status,
			"ms", time.Since(start).Milliseconds())
	})
}

// instrument zählt Anfragen pro Route (Muster, nicht Pfad!) und Status.
func instrument(reg *prometheus.Registry, next http.Handler) http.Handler {
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "shorturl_http_requests_total",
		Help: "HTTP-Anfragen nach Route und Status.",
	}, []string{"route", "code"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "shorturl_http_request_duration_seconds",
		Help:    "Bearbeitungsdauer nach Route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route"})
	reg.MustRegister(requests, duration)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		route := r.Pattern // vom ServeMux gesetzt, z. B. "GET /{code}"
		if route == "" {
			route = "unmatched"
		}
		requests.WithLabelValues(route, strconv.Itoa(rec.status)).Inc()
		duration.WithLabelValues(route).Observe(time.Since(start).Seconds())
	})
}

// recoverer fängt Panics ab: 500 statt abgerissener Verbindung.
func recoverer(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.ErrorContext(r.Context(), "panic", "path", r.URL.Path, "panic", v)
				httpError(w, http.StatusInternalServerError, "interner Fehler")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
