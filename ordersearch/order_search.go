package ordersearch

import (
	"context"
	"fmt"
)

const collection = "commerce-order-events"

type OrderUpdate struct {
	ID       string `json:"id"`
	OrderID  string `json:"order_id"`
	Kind     string `json:"kind"`
	Customer string `json:"customer"`
	Text     string `json:"text"`
}

type Match struct {
	ID       string         `json:"id"`
	Score    float64        `json:"score"`
	Metadata map[string]any `json:"metadata"`
}

type Embedder interface {
	Embed(context.Context, string) ([]float64, error)
}

type VectorStore interface {
	CreateCollection(context.Context, string, int) error
	Upsert(context.Context, string, []Vector) error
	Query(context.Context, string, []float64, int, map[string]any) ([]Match, error)
}

type Vector struct {
	ID       string         `json:"id"`
	Values   []float64      `json:"values"`
	Metadata map[string]any `json:"metadata"`
}

type Service struct {
	embedder Embedder
	vectors  VectorStore
}

func New(embedder Embedder, vectors VectorStore) *Service {
	return &Service{embedder: embedder, vectors: vectors}
}

func (s *Service) Index(ctx context.Context, updates []OrderUpdate) error {
	if len(updates) == 0 {
		return fmt.Errorf("at least one order update is required")
	}
	vectors := make([]Vector, 0, len(updates))
	for _, update := range updates {
		embedding, err := s.embedder.Embed(ctx, update.Text)
		if err != nil {
			return fmt.Errorf("embed update %s: %w", update.ID, err)
		}
		vectors = append(vectors, Vector{ID: update.ID, Values: embedding, Metadata: map[string]any{
			"order_id": update.OrderID, "kind": update.Kind, "customer": update.Customer, "text": update.Text,
		}})
	}
	if err := s.vectors.CreateCollection(ctx, collection, len(vectors[0].Values)); err != nil {
		return err
	}
	return s.vectors.Upsert(ctx, collection, vectors)
}

func (s *Service) Search(ctx context.Context, query, kind string, topK int) ([]Match, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if topK < 1 || topK > 20 {
		return nil, fmt.Errorf("top_k must be between 1 and 20")
	}
	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed search query: %w", err)
	}
	filter := map[string]any{}
	if kind != "" {
		filter["kind"] = kind
	}
	return s.vectors.Query(ctx, collection, embedding, topK, filter)
}
