package ordersearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type fixedEmbedder struct{ vector []float64 }

func (f fixedEmbedder) Embed(context.Context, string) ([]float64, error) { return f.vector, nil }

func TestSearchHandsEmbeddingToFilteredVectorQuery(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want map[string]any
	}{
		{name: "receipt lookup", kind: "receipt", want: map[string]any{"kind": "receipt"}},
		{name: "all order updates", kind: "", want: map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1/vector/query" {
					t.Fatalf("request = %s %s", r.Method, r.URL.Path)
				}
				var got struct {
					Embedding []float64      `json:"embedding"`
					Filter    map[string]any `json:"filter"`
				}
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got.Embedding, []float64{0.25, 0.75}) || !reflect.DeepEqual(got.Filter, tt.want) {
					t.Fatalf("handoff = %#v, %#v", got.Embedding, got.Filter)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"data":{"matches":[{"id":"receipt-1042","score":0.97,"metadata":{"kind":"receipt"}}]},"error":null,"metadata":{}}`))
			}))
			defer server.Close()

			store := &HTTPVectorStore{base: server.URL, apiKey: "test-key", http: server.Client()}
			service := New(fixedEmbedder{vector: []float64{0.25, 0.75}}, store)
			matches, err := service.Search(context.Background(), "where is my receipt", tt.kind, 3)
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) != 1 || matches[0].ID != "receipt-1042" {
				t.Fatalf("matches = %#v", matches)
			}
		})
	}
}

func TestEnvelopeErrorPreservesClientStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"code":"INVALID_ARGUMENT","message":"top_k is invalid"},"metadata":{}}`))
	}))
	defer server.Close()

	store := &HTTPVectorStore{base: server.URL, apiKey: "test-key", http: server.Client()}
	service := New(fixedEmbedder{vector: []float64{1}}, store)
	_, err := service.Search(context.Background(), "receipt", "receipt", 2)
	infraiErr, ok := err.(*InfraiError)
	if !ok || infraiErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("error = %#v", err)
	}
}
