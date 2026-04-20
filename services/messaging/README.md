# Messaging Service

**Message storage and real-time fan-out via Server-Sent Events.**

The Messaging service stores conversation history and broadcasts messages to connected clients in real-time. It is the source of truth for message content and metadata. Membership validation is delegated to the Channels service.

---

## What It Does

- **Message Storage** — Persist messages to PostgreSQL with metadata (channel, user, timestamp)
- **Message Retrieval** — Fetch history via cursor-based pagination
- **Real-Time Broadcast** — Push messages to all connected SSE subscribers in real-time
- **Membership Validation** — Enforce that only channel members can send messages

---

## HTTP API

All endpoints require JWT authentication (Bearer token), except `/stream` which accepts token via query parameter.

### Messages

#### POST `/api/v1/messages`

Send a message to a channel.

**Request Body:**
```json
{
  "channel_id": "550e8400-e29b-41d4-a716-446655440002",
  "content": "Hello team, here's an update on the project."
}
```

**Constraints:**
- `channel_id` — UUID of the channel (must exist)
- `content` — max 4000 characters

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "channel_id": "550e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Hello team, here's an update on the project.",
  "created_at": 1639747350
}
```

**Side Effects:**
1. Validates sender membership via **Channels service gRPC** (`ValidateMembership`)
2. Stores message in database
3. Broadcasts message via in-memory broker to all SSE subscribers
4. If validation fails, message is not sent (fail-closed)

**Error Responses:**
- `403 Forbidden` — Sender is not a member of the channel (`ErrNotAMember`)
- `503 Service Unavailable` — Channels service is unavailable (membership check failed)
- `400 Bad Request` — Validation error (missing fields, content too long)

---

#### GET `/api/v1/messages`

Fetch messages from a channel (cursor-based pagination, newest-first in DB; reversed for UI).

**Query Parameters:**
```
GET /api/v1/messages?channel_id=550e8400-e29b-41d4-a716-446655440002&limit=50&before_id=550e8400-e29b-41d4-a716-446655440010
```

| Param | Type | Default | Max | Description |
|-------|------|---------|-----|-------------|
| `channel_id` | UUID | required | — | Channel to fetch from |
| `before_id` | UUID | — | — | Cursor: message ID to fetch older messages from (exclusive); omit for newest messages |
| `limit` | int | 50 | 100 | Number of messages per page |

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "channel_id": "550e8400-e29b-41d4-a716-446655440002",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "content": "Hello team...",
      "created_at": 1639747350
    },
    ...
  ]
}
```

**Pagination Example:**
1. First call: `GET /messages?channel_id=...&limit=50` → returns messages 1–50 (newest)
2. Second call: `GET /messages?channel_id=...&limit=50&before_id=<id_of_message_50>` → returns messages 51–100 (older)
3. Repeat until `messages[]` is empty (no more history)

---

#### GET `/api/v1/messages/stream`

Open an SSE stream to receive messages from one or more channels in real-time.

**Query Parameters:**
```
GET /api/v1/messages/stream?channel_id=...&channel_id=...&token=<access_token>
```

| Param | Description |
|-------|-------------|
| `channel_id` | One or more channel UUIDs (repeat the param) to subscribe to; can specify 1+ channels |
| `token` | JWT access token (passed in query because EventSource cannot set custom headers) |

**Stream Events:**

```
event: message
data: {
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "channel_id": "550e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Hello team...",
  "created_at": 1639747350
}

: heartbeat
```

**Behavior:**
- On connect: immediately ready to receive messages
- On message sent: broadcasts to all subscribers on that channel (no ordering guarantee across channels)
- Every 15 seconds: sends a keepalive comment (`: heartbeat`)
- No snapshot of history is sent on connect; client must fetch via `GET /messages`

**Error Responses:**
- `401 Unauthorized` — Invalid token
- `400 Bad Request` — Missing `channel_id` parameter

**Multi-Channel Subscription:**
A single SSE connection can subscribe to multiple channels:

```
GET /api/v1/messages/stream?channel_id=ch1&channel_id=ch2&channel_id=ch3&token=...
```

This is more efficient than opening 3 separate connections.

---

## gRPC API

The gRPC service runs on port 50054. RPCs are defined in `proto/messaging.proto`.

| RPC | Request | Response |
|-----|---------|----------|
| `SendMessage` | `channel_id`, `user_id`, `content` | `Message` |
| `GetMessages` | `channel_id`, `before_id`, `limit` | `GetMessagesResponse` (`messages[]`) |

**Note:** These gRPC RPCs are used internally by other services (e.g., assist service fetches recent messages). The HTTP endpoints above are for client use.

---

## Database Schema

### `messages` Table

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,              -- max 4000 chars
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_channel_created ON messages(channel_id, created_at DESC);
```

**Index Note:** The composite index on `(channel_id, created_at DESC)` enables efficient history fetching: scan the channel's messages in reverse timestamp order (newest first).

---

## Environment Variables

All required at startup.

| Variable | Description | Example |
|----------|-------------|---------|
| `MESSAGING_DATABASE_URL` | PostgreSQL DSN | `postgres://postgres:password@localhost:5437/messaging?sslmode=disable` |
| `JWT_PUBLIC_KEY` | RS256 public key (PEM format) | (multi-line PEM content) |
| `MESSAGING_GRPC_PORT` | Port for gRPC server | `50054` |
| `MESSAGING_HTTP_PORT` | Port for HTTP server | `8084` |
| `CHANNELS_GRPC_ADDRESS` | Address of channels service | `localhost:50053` |

---

## Development

### Run Migrations

```bash
cd services/messaging
export DATABASE_URL="postgres://postgres:password@localhost:5437/messaging?sslmode=disable"
make migrate-up
```

### Run Service Locally

```bash
export MESSAGING_DATABASE_URL="postgres://postgres:password@localhost:5437/messaging?sslmode=disable"
export JWT_PUBLIC_KEY="$(cat ../../dev/public.pem)"
export MESSAGING_GRPC_PORT=50054
export MESSAGING_HTTP_PORT=8084
export CHANNELS_GRPC_ADDRESS="localhost:50053"

make run
```

### Run Tests

```bash
cd services/messaging
make test
```

---

## Cursor-Based Pagination

Unlike offset-based pagination, cursor-based pagination uses a stable reference point (a message ID) to fetch the next page:

**Advantages:**
- Efficient even with very large datasets (no offset scanning)
- Stable under concurrent inserts (offset pagination can skip/duplicate rows)
- Simpler to reason about ("fetch messages older than this one")

**Example Flow:**

```
Initial load (no cursor):
  GET /api/v1/messages?channel_id=ch1&limit=50
  ← Returns messages 1–50 (newest first)

Next page:
  GET /api/v1/messages?channel_id=ch1&limit=50&before_id=50
  ← Returns messages 51–100 (older)

Stop when:
  GET /api/v1/messages?channel_id=ch1&limit=50&before_id=100
  ← Returns 0 messages (reached end of history)
```

---

## Membership Enforcement

The messaging service enforces a **fail-closed** check: before accepting a message, it calls:

```
Channels.ValidateMembership(channel_id, user_id)
```

If the channels service is unavailable or if the user is not a member, the message is rejected with `403 Forbidden` or `503 Service Unavailable`. This prevents messages from being sent by non-members, even if it's inconvenient during an outage.

---

## Real-Time Fan-Out

Messages are pushed to clients via an in-memory broker:

1. **HTTP handler** receives message from client
2. **Service layer** validates, stores, and calls `broker.Publish(event)`
3. **Broker** distributes the event to all subscribedSSE connections (non-blocking)
4. **Slow clients** that can't keep up are silently dropped
5. **No message queue**: messages are only sent to clients currently connected

This design is simple and suitable for a POC. For production, consider a message queue (Kafka, RabbitMQ) to buffer events and decouple producers from consumers.

---

## Error Handling & Status Codes

| Error | HTTP Status | gRPC Code | Reason |
|-------|-------------|-----------|--------|
| `ErrChannelNotFound` | 404 | `NotFound` | Channel does not exist |
| `ErrNotAMember` | 403 | `PermissionDenied` | Sender is not a member of the channel |
| `ErrChannelsServiceUnavailable` | 503 | `Unavailable` | Channels service unreachable (membership check failed) |

---

## Inter-Service Communication

**Called By:**
- Assist service: `GetMessages` gRPC RPC to fetch recent message context for AI suggestions

**Calls:**
- Channels service: `ValidateMembership` gRPC RPC (fail-closed; blocks message send if unavailable)

---

## SSE Connection Limits

The service does not impose a hard limit on concurrent SSE connections. For production, consider:
- Setting a per-instance connection limit
- Implementing a load balancer that distributes SSE subscribers across instances
- Monitoring memory usage (each connection holds a goroutine + buffered channel)

---

## Message Ordering

Messages are delivered in the order they are inserted into the database (by `created_at`). However, when subscribing to multiple channels via a single SSE connection, **there is no global ordering**: messages from different channels are interleaved in the order they are published.

If strict global ordering is needed, switch to a single-channel-per-connection model.
