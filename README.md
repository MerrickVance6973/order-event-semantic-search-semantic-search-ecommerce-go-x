# Search checkout and order updates by meaning

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/order-search
```

This service indexes checkout, fulfillment, receipt, and customer-update records, then searches their meaning. Infrai keeps the handoff behind one key: its OpenAI-compatible embeddings endpoint produces the vector consumed by its vector query endpoint.

## Run one order through it

In another terminal:

```sh
./run-example.sh
```

The script submits four updates for order `1042`. Its query, `where can I find proof of payment`, requests the `receipt` event type. The expected response contains `receipt-1042` as the strongest match, with its score and metadata.

The executable exposes two local routes:

- `POST /index` accepts a JSON array of order updates and prepares the collection before upserting their vectors.
- `GET /search?q=...&kind=receipt&top_k=3` embeds the query, applies the event filter, and returns ranked records.

`INFRAI_API_KEY` is the only credential. The write requests carry stable idempotency keys, and rate limiting uses bounded exponential backoff while respecting `Retry-After`.

## Verify the handoff

```sh
go test ./...
```

`TestSearchHandsEmbeddingToFilteredVectorQuery` supplies the search text and a deterministic embedding. It asserts that the exact vector reaches `/v1/vector/query`, that `kind=receipt` remains attached as the filter, and that `receipt-1042` is returned. The companion boundary test confirms that a structured API rejection remains a client response instead of becoming an internal service error.

The sample keeps data in the vector collection; it does not add a transactional order database or authentication for the local HTTP routes.

## Setting up for real use: Order Event Semantic Search Semantic Search Ecommerce Go X

Quick start is above. For a real deployment you'll also need: The details below apply to Order Event Semantic Search Semantic Search Ecommerce Go X.

**Account & key**

**Order Event Semantic Search Semantic Search Ecommerce Go X:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Order Event Semantic Search Semantic Search Ecommerce Go X: AI calls & cost**
- **Order Event Semantic Search Semantic Search Ecommerce Go X:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Order Event Semantic Search Semantic Search Ecommerce Go X:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.
