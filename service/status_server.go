package service

import (
	"encoding/json"
	"log"
	"net/http"
)

// StartStatusServer starts a lightweight HTTP server to expose the poller's state.
func (p *Poller) StartStatusServer(addr string) {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := p.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	log.Printf("Status server listening on http://%s/status", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Status server failed: %v", err)
	}
}
