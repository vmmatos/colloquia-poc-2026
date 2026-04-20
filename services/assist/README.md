# Assist Service

**AI-powered reply suggestions powered by local Ollama.**

The Assist service generates contextual reply suggestions for users by analyzing recent channel messages and feeding them to a local LLM (Ollama). Suggestions are cached for 30 seconds to avoid redundant LLM calls.

---

## What It Does

- **Generate Suggestions** — Given a channel and partial message, return 3 short reply suggestions
- **Message Context** — Fetch recent messages to ground suggestions in conversation
- **LLM Inference** — Call Ollama with a prompt template; parse JSON suggestions
- **Caching** — Cache results by channel + input (30-second TTL) to reduce LLM load
- **Graceful Degradation** — If Ollama is unavailable, return empty suggestions (do not block message input)

---

## How It Works

```
Client POST /api/v1/assist/suggestions
  (channel_id="...", current_input="Hello ", message_limit=5)
        ↓
  Check in-memory cache by (channel_id + lowercase(input))
        ↓ [hit?] return cached suggestions
        ↓ [miss]
  Fetch last 5 messages from Messaging service gRPC
        ↓
  Build prompt:
    System: "You are a helpful assistant..."
    User: "[Last messages...]
           Complete this message: 'Hello '"
        ↓
  POST Ollama /api/generate (streaming NDJSON)
    ├ model: "qwen2.5:0.5b"
    ├ temperature: 0.6
    ├ top_k: 20
    ├ top_p: 0.9
    └ format: "json"
        ↓
  Parse response: extract JSON array of 3 strings
        ↓
  Cache result (TTL 30s)
        ↓
  Return suggestions to client
```

---

## HTTP API

All endpoints require JWT authentication (Bearer token).

#### POST `/api/v1/assist/suggestions`

Generate reply suggestions for a user's current input.

**Request Body:**
```json
{
  "channel_id": "550e8400-e29b-41d4-a716-446655440002",
  "current_input": "Hello team, I wanted to say ",
  "message_limit": 5
}
```

**Constraints:**
- `channel_id` — UUID of the channel
- `current_input` — max 500 characters; min 0 (empty input allowed, but suggestions will be generic)
- `message_limit` — how many recent messages to fetch for context; default 5, min 0, max 50

**Response (200 OK):**
```json
{
  "suggestions": [
    "that we're making great progress on the project.",
    "that I've completed the code review.",
    "that the tests are now passing."
  ]
}
```

Always returns exactly 3 suggestions (or fewer if fewer are available).

**Error Responses:**
- `400 Bad Request` — Validation error (missing fields, input too long)
- `500 Internal Server Error` — LLM inference failed (extremely rare; service should gracefully degrade)

**Timeout:** Gateway timeout is 70 seconds to accommodate LLM inference latency (configurable via `ASSIST_OLLAMA_TIMEOUT_SECONDS`).

---

## gRPC API

The gRPC service runs on port 50055. RPCs are defined in `proto/assist.proto`.

| RPC | Request | Response |
|-----|---------|----------|
| `GetSuggestions` | `channel_id`, `current_input`, `message_limit` | `GetSuggestionsResponse` (`suggestions[]`) |

---

## Caching

Suggestions are cached in-memory with a **30-second TTL**:

**Cache Key:** `channel_id + "|" + lowercase(trimmed(current_input))`

**Example:**
- Input: `"Hello "` (with trailing space)
- Cached as: `"550e8400.../hello"`
- If user types `"HELLO "` later, reuses the same cache entry (case-insensitive)

**Expiry:** Lazy expiry on read — when a cache entry is retrieved, check if it's older than 30 seconds; if so, delete and recompute.

**Implementation:** `sync.Map` for concurrent-safe reads without locking.

---

## LLM Configuration

### Ollama Setup

**Base URL:** Controlled by `ASSIST_OLLAMA_BASE_URL` (e.g., `http://ollama:11434`)

**Model:** Specified by `ASSIST_OLLAMA_MODEL` (e.g., `qwen2.5:0.5b`)

### Inference Parameters

| Parameter | Value | Reason |
|-----------|-------|--------|
| `temperature` | 0.6 | Balanced creativity vs consistency (0=deterministic, 1=random) |
| `top_k` | 20 | Consider top 20 most likely tokens |
| `top_p` | 0.9 | Nucleus sampling; consider tokens until cumulative probability reaches 90% |
| `repeat_penalty` | 1.0 | No penalty for repeating tokens (default behavior) |
| `num_predict` | 256 | Max 256 tokens per completion (hard limit to prevent rambling) |
| `num_ctx` | 1024 | Context window size (1024 tokens) |

### JSON-Constrained Output

The `/api/generate` request includes `"format": "json"`, which instructs Ollama to constrain the output to valid JSON. This ensures the response can be parsed as JSON without extraction heuristics.

### Warm-Up on Startup

On service startup, a warm-up request is fired (timeout 120 seconds) to pre-load the model into Ollama's memory. This prevents the first user request from incurring the full model-loading latency (which can be several seconds).

---

## Prompt Template

**System Prompt (fixed):**
```
You are a helpful assistant that completes messages in a team chat application. 
You understand the context of the conversation and complete messages naturally, 
as if continuing a human's thought.

Your task is to generate exactly 3 short and plausible completions for the given message.
Return them as a JSON array of strings, nothing else.

Example:
[
  "that I've made good progress on the task",
  "that the deadline has been extended",
  "that I need feedback from the team"
]
```

**User Prompt (dynamic):**
```
Context from the conversation:
[Recent messages, each truncated at 200 characters]

Complete this message: "<current_input>"
```

---

## Environment Variables

All required at startup.

| Variable | Description | Example |
|----------|-------------|---------|
| `ASSIST_HTTP_PORT` | Port for HTTP server | `8085` |
| `ASSIST_GRPC_PORT` | Port for gRPC server | `50055` |
| `JWT_PUBLIC_KEY` | RS256 public key (PEM format) | (multi-line PEM content) |
| `ASSIST_MESSAGING_GRPC_ADDRESS` | Address of messaging service | `localhost:50054` |
| `ASSIST_LLM_PROVIDER` | LLM backend selector | `"ollama"` (only valid value) |
| `ASSIST_OLLAMA_BASE_URL` | Ollama base URL | `http://ollama:11434` |
| `ASSIST_OLLAMA_MODEL` | Model name | `qwen2.5:0.5b` |
| `ASSIST_OLLAMA_TIMEOUT_SECONDS` | Per-request inference timeout | `60` |

---

## Development

### Set Up Ollama Locally

1. **Install Ollama:** https://ollama.ai
2. **Start Ollama:** `ollama serve` (default port 11434)
3. **Pull a model:** `ollama pull qwen2.5:0.5b` (lightweight, ~400MB)

### Run Service Locally

```bash
export ASSIST_HTTP_PORT=8085
export ASSIST_GRPC_PORT=50055
export JWT_PUBLIC_KEY="$(cat ../../dev/public.pem)"
export ASSIST_MESSAGING_GRPC_ADDRESS="localhost:50054"
export ASSIST_LLM_PROVIDER="ollama"
export ASSIST_OLLAMA_BASE_URL="http://localhost:11434"
export ASSIST_OLLAMA_MODEL="qwen2.5:0.5b"
export ASSIST_OLLAMA_TIMEOUT_SECONDS=60

make run
```

### Run Tests

```bash
cd services/assist
make test
```

---

## Model Selection

The service uses **Ollama** as the LLM provider. To swap models:

1. Pull a different model: `ollama pull <model_name>`
2. Set `ASSIST_OLLAMA_MODEL=<model_name>`
3. Restart the service

**Recommended models for text completion:**
- `qwen2.5:0.5b` — Very fast, small (400 MB)
- `qwen2.5:1.5b` — Balanced speed/quality (900 MB)
- `llama2:7b` — Higher quality, slower (4 GB)
- `mistral:7b` — Fast, good quality (4 GB)

---

## Error Handling

If Ollama is unavailable or times out:
- The service returns `{ "suggestions": [] }` (empty array)
- Error is logged but does not block the user
- Client displays empty suggestions state

This "graceful degradation" approach ensures that the assist feature being offline does not break the messaging experience.

---

## Performance Considerations

### Cache Hit Rate

With typical usage (users are indecisive, typing the same input multiple times), cache hit rates of 70–90% are common, reducing LLM load significantly.

### Per-Request Latency

- **Cache hit:** <10ms
- **Cache miss + LLM inference:** 3–10 seconds (depends on model size and input length)
- **Gateway timeout:** 70 seconds (very conservative)

### Throughput

A single Ollama instance on modern hardware (e.g., MacBook M2) can handle ~1–2 concurrent inference requests. For higher throughput:
- Run multiple Ollama instances and load-balance
- Use a larger model queue/context window
- Use a faster model (e.g., `qwen2.5:0.5b`)

---

## Suggestions Quality

Suggestion quality depends on:
1. **Model choice** — larger models (7B+) produce more coherent suggestions
2. **Context length** — more recent messages = better suggestions
3. **Input length** — longer partial input = more specific suggestions

For a POC, the small `qwen2.5:0.5b` model is perfectly adequate. For production, test with larger models (7B+) and measure user satisfaction.

---

## Future Enhancements

- **User preferences** — Track which suggestions users accept; fine-tune model
- **Channel context** — Include channel description/topic in system prompt
- **User profile** — Condition suggestions on user's communication style (from past messages)
- **Batch requests** — Cache suggestions for multiple inputs in parallel
- **Custom models** — Allow orgs to swap the LLM backend (e.g., local GPT, OpenAI API)

---

## Inter-Service Communication

**Called By:**
- Frontend: HTTP POST to suggestions endpoint

**Calls:**
- Messaging service: `GetMessages` gRPC RPC to fetch recent message context (best-effort; continues with empty context if unavailable)
- Ollama: HTTP POST `/api/generate` (best-effort; returns empty suggestions if unavailable)
