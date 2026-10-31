// Package server stellt die HTTP-API von shorturl bereit.
package server

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"example.com/shorturl/internal/store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var validCode = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)

type server struct {
	store   store.Store
	log     *slog.Logger
	baseURL string
}

// New baut den Handler mit allen Routen und Middleware.
func New(st store.Store, log *slog.Logger, baseURL string,
	reg *prometheus.Registry) http.Handler {
	s := &server{store: st, log: log, baseURL: baseURL}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/links", s.create)
	mux.HandleFunc("GET /api/links/{code}", s.get)
	mux.HandleFunc("DELETE /api/links/{code}", s.delete)
	mux.HandleFunc("GET /{code}", s.redirect)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return recoverer(log, instrument(reg, logRequests(log, mux)))
}

type createReq struct {
	URL  string `json:"url"`
	Code string `json:"code,omitempty"` // optional, sonst zufällig
}

type createResp struct {
	store.Link
	Short string `json:"short"`
}

func (s *server) create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "ungültiges JSON: "+err.Error())
		return
	}
	if u, err := url.ParseRequestURI(req.URL); err != nil || u.Host == "" ||
		(u.Scheme != "http" && u.Scheme != "https") {
		httpError(w, http.StatusBadRequest, "url muss http(s)://host... sein")
		return
	}
	if req.Code != "" && (!validCode.MatchString(req.Code) || reserved(req.Code)) {
		httpError(w, http.StatusBadRequest, "code: 3-32 Zeichen A-Z a-z 0-9 _ -")
		return
	}

	link := store.Link{Code: req.Code, URL: req.URL, Created: time.Now().UTC()}
	var err error
	for range 5 { // zufälliger Code kollidiert selten – dann neu würfeln
		if req.Code == "" {
			link.Code = randomCode()
		}
		err = s.store.Create(r.Context(), link)
		if req.Code != "" || !errors.Is(err, store.ErrExists) {
			break
		}
	}
	switch {
	case errors.Is(err, store.ErrExists):
		httpError(w, http.StatusConflict, "code existiert bereits")
	case err != nil:
		s.internal(w, r, err)
	default:
		writeJSON(w, http.StatusCreated, createResp{link, s.baseURL + "/" + link.Code})
	}
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	link, err := s.store.Get(r.Context(), r.PathValue("code"))
	if s.handleErr(w, r, err) {
		return
	}
	writeJSON(w, http.StatusOK, link)
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	err := s.store.Delete(r.Context(), r.PathValue("code"))
	if s.handleErr(w, r, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) redirect(w http.ResponseWriter, r *http.Request) {
	link, err := s.store.Hit(r.Context(), r.PathValue("code"))
	if s.handleErr(w, r, err) {
		return
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}

// handleErr schreibt 404 bzw. 500 und meldet, ob ein Fehler vorlag.
func (s *server) handleErr(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, store.ErrNotFound):
		httpError(w, http.StatusNotFound, "unbekannter code")
	default:
		s.internal(w, r, err)
	}
	return true
}

// internal loggt Details, der Client sieht nur eine allgemeine Meldung.
func (s *server) internal(w http.ResponseWriter, r *http.Request, err error) {
	s.log.ErrorContext(r.Context(), "interner Fehler", "path", r.URL.Path, "err", err)
	httpError(w, http.StatusInternalServerError, "interner Fehler")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func reserved(code string) bool {
	return code == "api" || code == "healthz" || code == "metrics"
}

const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// randomCode: 7 Zeichen ohne Verwechsler (l/1, O/0) aus crypto/rand.
func randomCode() string {
	b := make([]byte, 7)
	rand.Read(b) // liefert nie einen Fehler (Go 1.24+)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}
