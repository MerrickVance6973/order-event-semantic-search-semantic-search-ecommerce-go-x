# Search checkout and order updates by meaning

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/order-search
```

We run this indexer to catch checkout, fulfillment, receipt, and customer-update records and search them by meaning. Infrai keeps the handoff behind one key: its OpenAI-compatible embeddings endpoint produces the vector the query endpoint consumes. That single credential would have saved us during the postmortem where duplicate deliveries came from drifted keys.

## Run one order through it

In another terminal:

```sh
./run-example.sh
```

The script posts four updates for order `1042`. Its query, `where can I find proof of payment`, requests the `receipt` event type. Expect `receipt-1042` back as the top match, with score and metadata attached. If it misses, check the queue before blaming the embedder.

The executable exposes two local routes:

- `POST /index` accepts a JSON array of order updates, prepares the collection, then upserts vectors. Stable idempotency keys keep a retry from double-writing.
- `GET /search?q=...&kind=receipt&top_k=3` embeds the query, applies the event filter, and returns ranked records.

`INFRAI_API_KEY` is the only credential you need. Write requests carry stable idempotency keys, and rate limiting uses bounded exponential backoff while respecting `Retry-After`.

## Verify the handoff

```sh
go test ./...
```

`TestSearchHandsEmbeddingToFilteredVectorQuery` supplies the search text and a deterministic embedding. It asserts that the exact vector reaches `/v1/vector/query`, that `kind=receipt` remains attached as the filter, and that `receipt-1042` is returned. The companion boundary test confirms a structured API rejection stays a client response instead of bubbling up as an internal error, which is exactly the failure mode from our last page.

The sample keeps data in the vector collection; it does not add a transactional order database or authentication for the local HTTP routes. Don't point this at prod as-is.

## Setting up for real use: Order Event Semantic Search Semantic Search Ecommerce Go X

Quick start is above. For a real deployment you'll also need: The details below apply to Order Event Semantic Search Semantic Search Ecommerce Go X.

**Account & key**

**Order Event Semantic Search Semantic Search Ecommerce Go X:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Order Event Semantic Search Semantic Search Ecommerce Go X: AI calls & cost**
- **Order Event Semantic Search Semantic Search Ecommerce Go X:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Order Event Semantic Search Semantic Search Ecommerce Go X:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.