package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type AIRequest struct {
	Text string `json:"text"`
}

type AIResponse struct {
	Reply string                 `json:"reply"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
	Action string                `json:"action,omitempty"`
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func forwardToAI(aiURL string, text string) (*AIResponse, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	reqBody := AIRequest{Text: text}
	b, _ := json.Marshal(reqBody)
	resp, err := client.Post(aiURL+"/api/process", "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, err
	}
	return &aiResp, nil
}

func commandHandler(aiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		aiResp, err := forwardToAI(aiURL, req.Text)
		if err != nil {
			http.Error(w, "failed to contact AI service: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(aiResp)
	}
}

func main() {
	aiURL := getEnv("PY_AI_URL", "http://localhost:5000")
	httpPort := getEnv("PORT", "8080")

	http.HandleFunc("/api/command", commandHandler(aiURL))

	fmt.Println("Go Orchestrator listening on :8080, forwarding to", aiURL)
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
		log.Fatal(err)
	}
}
