package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	mu             sync.Mutex
	description    string
	image          string
	AnalyzeHandler http.HandlerFunc
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start(addr string) error {
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/observation", s.handleObservation)
	if s.AnalyzeHandler != nil {
		http.HandleFunc("/api/analyze", s.AnalyzeHandler)
	}

	log.Printf("Web server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleObservation(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	desc := s.description
	img := s.image
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"image":       img,
		"description": desc,
	})
}

func (s *Server) UpdateObservation(img string, description string) {
	s.mu.Lock()
	s.image = img
	s.description = description
	s.mu.Unlock()
}
