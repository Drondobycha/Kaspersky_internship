package server

import (
	"Kaspersky_internship/internal/config"
	"Kaspersky_internship/internal/queue"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	queue      *queue.Queue
	config     *config.Config
	shutdown   chan struct{}
	wg         sync.WaitGroup
}

type EnqueueRequest struct {
	ID         string `json:"id"`
	Payload    string `json:"payload"`
	MaxRetries int    `json:"max_retries"`
}

type EnqueueResponse struct {
	Status  string `json:"status"`
	ID      string `json:"id"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status    string           `json:"status"`
	Timestamp string           `json:"timestamp"`
	Queue     queue.QueueStats `json:"queue"`
}

func NewServer(cfg *config.Config, taskHandler queue.TaskHandler) *Server {
	q := queue.NewQueue(cfg.QueueBufferSize, taskHandler)

	mux := http.NewServeMux()
	s := &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		queue:    q,
		config:   cfg,
		shutdown: make(chan struct{}),
	}

	// Register handlers
	mux.HandleFunc("POST /enqueue", s.handleEnqueue)
	mux.HandleFunc("GET /healthz", s.handleHealthCheck)

	return s
}

func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	if s.queue.IsStopped() {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "Field 'id' is required", http.StatusBadRequest)
		return
	}
	if req.Payload == "" {
		http.Error(w, "Field 'payload' is required", http.StatusBadRequest)
		return
	}
	if req.MaxRetries < 0 {
		http.Error(w, "Field 'max_retries' must be non-negative", http.StatusBadRequest)
		return
	}

	if req.MaxRetries == 0 {
		req.MaxRetries = s.config.MaxRetries
	}

	task := &queue.Task{
		ID:         req.ID,
		Payload:    req.Payload,
		MaxRetries: req.MaxRetries,
	}

	if err := s.queue.EnQueue(task); err != nil {
		http.Error(w, fmt.Sprintf("Failed to enqueue task: %v", err), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(EnqueueResponse{
		Status:  "accepted",
		ID:      task.ID,
		Message: "Task added to queue",
	})
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	stats := s.queue.GetStats()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Queue:     stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) Start() error {
	// Start workers
	ctx := context.Background()
	s.queue.StartWorkers(ctx, s.config.WorkersCount)

	// Start HTTP server
	go func() {
		log.Printf("Starting server on %s", s.config.ServerAddr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Handle shutdown signals
	return s.waitForShutdown()
}

func (s *Server) waitForShutdown() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
	case <-s.shutdown:
		log.Printf("Shutdown requested")
	}

	return s.Stop()
}

func (s *Server) Stop() error {
	log.Printf("Stopping queue...")
	s.queue.Stop()

	log.Printf("Stopping HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		return err
	}

	log.Printf("Server stopped gracefully")
	return nil
}

func (s *Server) GracefulStop() {
	close(s.shutdown)
}
