package api

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
)

type usageResponse struct {
	TotalTokens         int64 `json:"total_tokens"`
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	CacheReadTokens     int64 `json:"cache_read_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_tokens"`
}

type transcriptMessage struct {
	Type    string `json:"type"`
	Message struct {
		Usage struct {
			InputTokens              int64 `json:"input_tokens"`
			OutputTokens             int64 `json:"output_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

func (s *Server) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	path := s.db.GetSessionTranscriptPath(id)
	if path == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(usageResponse{})
		return
	}

	f, err := os.Open(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(usageResponse{})
		return
	}
	defer f.Close()

	var resp usageResponse
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		var msg transcriptMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		if msg.Type != "assistant" {
			continue
		}
		u := msg.Message.Usage
		resp.InputTokens += u.InputTokens
		resp.OutputTokens += u.OutputTokens
		resp.CacheReadTokens += u.CacheReadInputTokens
		resp.CacheCreationTokens += u.CacheCreationInputTokens
	}

	resp.TotalTokens = resp.InputTokens + resp.OutputTokens + resp.CacheReadTokens + resp.CacheCreationTokens

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
