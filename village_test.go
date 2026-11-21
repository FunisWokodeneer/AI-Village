package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVillageRunKeepsAgentsInSpecialties(t *testing.T) {
	var received []ollamaRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request ollamaRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		received = append(received, request)
		_ = json.NewEncoder(w).Encode(ollamaResponse{
			Message: OllamaMessage{Role: "assistant", Content: "specialized result"},
		})
	}))
	defer server.Close()

	village := NewVillage("test-model", server.URL)
	results, err := village.Run(context.Background(), "create a project")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(results) != len(Agents) {
		t.Fatalf("got %d results, want %d", len(results), len(Agents))
	}
	if len(received) != len(Agents) {
		t.Fatalf("got %d requests, want %d", len(received), len(Agents))
	}
	for i, request := range received {
		if len(request.Messages) != 2 {
			t.Fatalf("agent %d got %d messages", i, len(request.Messages))
		}
		if !strings.Contains(request.Messages[0].Content, "Only") {
			t.Fatalf("agent %d prompt does not enforce its specialty", i)
		}
	}
}

func TestVillageRunRejectsEmptyGoal(t *testing.T) {
	village := NewVillage("test-model", "http://127.0.0.1")
	if _, err := village.Run(context.Background(), "  "); err == nil {
		t.Fatal("Run accepted an empty goal")
	}
}
