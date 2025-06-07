package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Request structures for Copilot extension
type CopilotRequest struct {
	Messages []Message `json:"messages"`
	User     User      `json:"user"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type User struct {
	Login string `json:"login"`
}

// Response structures
type CopilotResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index int     `json:"index"`
	Delta Delta   `json:"delta"`
	Finish string `json:"finish_reason,omitempty"`
}

type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type GreetingResponse struct {
	Message string `json:"message"`
}

// GET endpoint for greeting
func greetingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	response := GreetingResponse{
		Message: "Hello! I'm your table cache extension. Ask me about table cache with @myextension what's the table cache of [TABLE_NAME]",
	}
	
	json.NewEncoder(w).Encode(response)
}

// POST endpoint for chat completions (Copilot extension format)
func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var req CopilotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get the last user message
	var userMessage string
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			userMessage = msg.Content
		}
	}

	log.Printf("Received message: %s", userMessage)

	// Determine response type based on message content
	var responseContent string
	if isGreetingMessage(userMessage) {
		responseContent = getGreetingMessage()
	} else if isTableCacheQuery(userMessage) {
		tableName := extractTableName(userMessage)
		responseContent = generateTableCacheResponse(tableName)
	} else {
		responseContent = getHelpMessage()
	}

	// Set headers for Server-Sent Events (SSE) format that Copilot expects
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send the response in Copilot's expected format
	response := CopilotResponse{
		Choices: []Choice{
			{
				Index: 0,
				Delta: Delta{
					Role:    "assistant",
					Content: responseContent,
				},
				Finish: "stop",
			},
		},
	}

	// Convert to JSON and send
	jsonData, _ := json.Marshal(response)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
}

// Check if message is a greeting
func isGreetingMessage(message string) bool {
	message = strings.ToLower(strings.TrimSpace(message))
	greetingWords := []string{"hi", "hello", "hey", "greeting", "greet"}
	
	for _, word := range greetingWords {
		if strings.Contains(message, word) {
			return true
		}
	}
	return false
}

// Check if message is asking about table cache
func isTableCacheQuery(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "table cache") || 
		   strings.Contains(message, "cache of") ||
		   strings.Contains(message, "cache for")
}

// Get greeting message
func getGreetingMessage() string {
	return `Hello! 👋 I'm your Table Cache Extension!

I can help you get information about database table caches. Here's how to use me:

**Available Commands:**
• Just say "hi" or "hello" for this greeting
• Ask "what's the table cache of [TABLE_NAME]" to get cache details

**Example:**
• @my-table-cache-copilot what's the table cache of TBCD
• @my-table-cache-copilot what's the table cache of USERS

**Available Tables:** TBCD, USERS, ORDERS

What would you like to know about your table caches?`
}

// Get help message for unknown queries
func getHelpMessage() string {
	return `I'm not sure what you're asking for. Here's what I can help you with:

**Available Commands:**
• Say "hi" or "hello" for a greeting
• Ask "what's the table cache of [TABLE_NAME]" to get cache information

**Examples:**
• @my-table-cache-copilot hi
• @my-table-cache-copilot what's the table cache of TBCD

**Available Tables:** TBCD, USERS, ORDERS

Please try one of these commands!`
}
func extractTableName(message string) string {
	// Look for patterns like "table cache of TBCD" or "TBCD table cache"
	message = strings.ToLower(message)
	
	// Remove common words and find the table name
	words := strings.Fields(message)
	for i, word := range words {
		if word == "of" && i+1 < len(words) {
			return strings.ToUpper(words[i+1])
		}
	}
	
	// If no "of" pattern, look for uppercase words in original message
	originalWords := strings.Fields(strings.TrimSpace(message))
	for _, word := range originalWords {
		if len(word) > 2 && strings.ToUpper(word) == word {
			return word
		}
	}
	
	return "UNKNOWN"
}

// Generate table cache response
func generateTableCacheResponse(tableName string) string {
	switch tableName {
	case "TBCD":
		return `Table Cache Information for TBCD:
- MENUCACHE
- APICACHE
- TRANSCACHE`
	
	case "USERS":
		return `Table Cache Information for USERS:
- USERCACHE
- REGISTRYCACHE`
	
	case "ORDERS":
		return `Table Cache Information for ORDERS:
- ORDERCACHE`
	
	default:
		return fmt.Sprintf(`Table Cache Information for %s:
- Status: Table not found in cache system
- Suggestion: Please check if the table name is correct
- Available cached tables: TBCD, USERS, ORDERS
- Contact admin if you need to add this table to cache monitoring`, tableName)
	}
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "copilot-extension-api",
	})
}

func main() {
	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup routes
	http.HandleFunc("/", greetingHandler)
	http.HandleFunc("/greeting", greetingHandler)
	http.HandleFunc("/v1/chat/completions", chatHandler)  // Copilot extension endpoint
	http.HandleFunc("/health", healthHandler)

	// CORS middleware for all requests
	http.HandleFunc("/cors-proxy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		// Handle the actual request
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	log.Printf("Starting server on port %s", port)
	log.Printf("Endpoints available:")
	log.Printf("  GET  /greeting - Get greeting message")
	log.Printf("  POST /v1/chat/completions - Chat completions for Copilot")
	log.Printf("  GET  /health - Health check")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}