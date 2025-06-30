package main

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	b *Broker
}

func (s *Server) handleEventsStream(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id, messageChan := s.b.AddSubscriber()
	defer s.b.RemoveSubscriber(id)

	for {
		select {
		case html := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", html)
			flusher.Flush()
		case <-req.Context().Done():
			return
		}
	}
}

// Here we want to start the Server.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fileServer)
	mux.HandleFunc("/events", s.handleEventsStream)

	server := &http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	log.Println("Starting server on :8090")
	return server.ListenAndServe()
}
