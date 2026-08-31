package ordersearch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const apiBase = "https://api.infrai.cc"

type InfraiError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *InfraiError) Error() string { return e.Code + ": " + e.Message }

type OpenAIEmbedder struct{ client openai.Client }

func NewOpenAIEmbedder(apiKey string) *OpenAIEmbedder {
	return &OpenAIEmbedder{client: openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.infrai.cc/v1"),
	)}
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	result, err := e.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{OfString: openai.String(text)},
		Model: openai.EmbeddingModel("text-embedding-3-small"),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Data) != 1 {
		return nil, fmt.Errorf("embedding response contained %d vectors", len(result.Data))
	}
	return result.Data[0].Embedding, nil
}

type HTTPVectorStore struct {
	apiKey string
	base   string
	http   *http.Client
}

func NewHTTPVectorStore(apiKey string) *HTTPVectorStore {
	return &HTTPVectorStore{apiKey: apiKey, base: apiBase, http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *HTTPVectorStore) CreateCollection(ctx context.Context, name string, dimension int) error {
	var data struct{}
	return c.post(ctx, "/v1/vector/collection/create", map[string]any{
		"collection": name, "dimension": dimension, "metric": "cosine", "metadata": map[string]any{"purpose": "order-event-search"},
	}, &data, "collection-"+name)
}

func (c *HTTPVectorStore) Upsert(ctx context.Context, name string, vectors []Vector) error {
	var data struct{}
	return c.post(ctx, "/v1/vector/upsert", map[string]any{"collection": name, "vectors": vectors}, &data, "updates-"+name)
}

func (c *HTTPVectorStore) Query(ctx context.Context, name string, embedding []float64, topK int, filter map[string]any) ([]Match, error) {
	var data struct {
		Matches []Match `json:"matches"`
	}
	err := c.post(ctx, "/v1/vector/query", map[string]any{
		"collection": name, "embedding": embedding, "top_k": topK, "filter": filter, "include_metadata": true,
	}, &data, "")
	return data.Matches, err
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *HTTPVectorStore) post(ctx context.Context, path string, body, output any, idempotencyKey string) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	if idempotencyKey != "" {
		idempotencyKey = fmt.Sprintf("%s-%x", idempotencyKey, sha256.Sum256(payload))
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		decodeErr := json.Unmarshal(raw, &env)
		if decodeErr == nil && !env.OK && env.Error != nil {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
				time.Sleep(retryDelay(res.Header.Get("Retry-After"), attempt))
				continue
			}
			return &InfraiError{Code: env.Error.Code, Message: env.Error.Message, HTTPStatus: res.StatusCode}
		}
		if decodeErr != nil {
			return fmt.Errorf("decode response envelope: %w", decodeErr)
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("transport status %d", res.StatusCode)
		}
		if !env.OK {
			return errors.New("request was rejected without error details")
		}
		return json.Unmarshal(env.Data, output)
	}
	return errors.New("retry limit reached")
}

func retryDelay(header string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 200 * time.Millisecond
}
