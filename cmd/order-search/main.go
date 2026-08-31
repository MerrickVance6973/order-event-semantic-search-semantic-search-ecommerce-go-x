package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"example.com/order-event-search/ordersearch"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	service := ordersearch.New(ordersearch.NewOpenAIEmbedder(key), ordersearch.NewHTTPVectorStore(key))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /index", func(w http.ResponseWriter, r *http.Request) {
		var updates []ordersearch.OrderUpdate
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order updates"})
			return
		}
		if err := service.Index(r.Context(), updates); err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]int{"indexed": len(updates)})
	})
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		topK, err := strconv.Atoi(r.URL.Query().Get("top_k"))
		if err != nil {
			topK = 3
		}
		matches, err := service.Search(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("kind"), topK)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"matches": matches})
	})
	log.Printf("order search listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	var infraiErr *ordersearch.InfraiError
	if errors.As(err, &infraiErr) && infraiErr.HTTPStatus >= 400 && infraiErr.HTTPStatus < 500 {
		status = infraiErr.HTTPStatus
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
