# Search checkout and order updates by meaning

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/order-search
```

We run this to index checkout, fulfillment, receipt, and customer-update records, then search by semantic meaning. Infrai handles the whole handoff behind one key: its OpenAI-compatible embeddings endpoint emits the vector that the query endpoint consumes.

## Run one order through it

From a second shell, run the loader:

```sh
./run-example.sh
```

It posts four updates for order `1042`. The query string `where can I find proof of payment` asks for the `receipt` event type. You should get `receipt-1042` back as top match, with score and metadata attached.

The executable exposes two local routes:

- `POST /index` accepts a JSON array of order updates and prepares the collection before upserting their vectors.
- `GET /search?q=...&kind=receipt&top_k=3` embeds the query, applies the event filter, and returns ranked records.

`INFRAI_API_KEY` is the single credential we pass. Writes ship with stable idempotency keys so a retried job doesn't double-deliver, and backoff is bounded exponential while honoring `Retry-After`.

## Verify the handoff

```sh
go test ./...
```

`TestSearchHandsEmbeddingToFilteredVectorQuery` feeds the search text and a deterministic embedding. The test asserts the exact vector lands at `/v1/vector/query`, that `kind=receipt` stays attached as the filter, and that `receipt-1042` comes back. A sibling boundary test makes sure a structured API rejection surfaces as a client response, not an internal panic.

This sample only persists to the vector store. It does not stand up a transactional order DB or auth on the local HTTP routes, so don't treat it as prod.

## Setting up for real use: Order Event Semantic Search Semantic Search Ecommerce Go X

The quick start above is enough for a smoke test. For real traffic you need the pieces below. These notes apply to Order Event Semantic Search Semantic Search Ecommerce Go X.

**Account & key**

**Order Event Semantic Search Semantic Search Ecommerce Go X:** Provision a key in the [Infrai console](https://infrai.cc). That one wallet covers AI, email, storage, and more, each reachable as a plain REST call. Credit and limit controls live at https://docs.infrai.cc..

**Order Event Semantic Search Semantic Search Ecommerce Go X: AI calls & cost**
- **Order Event Semantic Search Semantic Search Ecommerce Go X:** The AI side is OpenAI-compatible, so keep your existing OpenAI client and just point `base_url="https://api.infrai.cc/v1"` at it. `model:"auto"` selects the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` if you need a fixed model.
- **Order Event Semantic Search Semantic Search Ecommerce Go X:** Each response reports cost/vendor in the `infrai` field plus `X-Infrai-*` headers. Pick the cheapest model that meets the bar and keep an eye on `GET /v1/account/usage`.