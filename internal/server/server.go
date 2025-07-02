package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
	"transport-nsw-exporter/pkg/collectors"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ready atomic.Value

type Server struct {
	listenAddress string
	logger        *slog.Logger
}

func New(listenAddress string, logger *slog.Logger) *Server {
	return &Server{
		listenAddress: listenAddress,
		logger:        logger,
	}
}

func (s *Server) Run() error {
	ready.Store(false)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler)
	mux.Handle("/metrics", promhttp.Handler())

	collectors.RegisterCollectors(s.logger)

	server := &http.Server{
		Addr:    s.listenAddress,
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		ready.Store(true)
		s.logger.Info("starting server", "web.listen-address", s.listenAddress)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-quit:
		s.logger.Info("shutdown signal received")
	case err := <-errChan:
		return err
	}

	ready.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	s.logger.Info("server stopped gracefully")
	return nil
}

// implement some actual logic here for liveness
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// implement some actual logic for readiness
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	if val, ok := ready.Load().(bool); ok && val {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
	}
}
