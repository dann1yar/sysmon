package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sysmon/internal/metrics"
)

type snapshotStore struct {
	mu   sync.Mutex
	snap Snapshot
	// false until the first sample pair is in, so an early scrape gets
	// a 503 instead of a snapshot full of zeros
	ready bool
}

func (s *snapshotStore) Set(snap Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snap = snap
	s.ready = true
}

func (s *snapshotStore) Get() (Snapshot, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snap, s.ready
}

func sampleLoop(ctx context.Context, store *snapshotStore, diskPath string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	prev := metrics.Take()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		cur := metrics.Take()
		store.Set(buildSnapshot(prev, cur, interval.Seconds(), diskPath))
		prev = cur
	}
}

func runServe(addr, diskPath string, interval time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &snapshotStore{}
	go sampleLoop(ctx, store, diskPath, interval)

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		snap, ready := store.Get()
		if !ready {
			http.Error(w, "sysmon: warming up, try again shortly", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		writePrometheus(w, snap)
	})
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		snap, ready := store.Get()
		if !ready {
			http.Error(w, "sysmon: warming up, try again shortly", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snap)
	})

	srv := &http.Server{Addr: addr, Handler: mux}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		srv.Shutdown(shutdownCtx)
	}()

	fmt.Printf("sysmon: слушаю на http://%s (/metrics, /json), Ctrl+C для выхода\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, "sysmon:", err)
		os.Exit(1)
	}
}
