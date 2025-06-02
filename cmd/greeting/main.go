package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type CopilotRequest struct {
	Prompt string `json:"prompt"`
}

type CopilotResponse struct {
	Content string `json:"content"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	var req CopilotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := CopilotResponse{
		Content: "Hello " + req.Prompt,
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/greet", handler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
