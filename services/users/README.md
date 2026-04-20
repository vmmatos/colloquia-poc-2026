# Users Service

**User profiles, real-time presence tracking, and search.**

The Users service stores user profile data (name, avatar, bio, timezone, language) and tracks which users are currently online. Real-time presence is broadcast via Server-Sent Events (SSE); a heartbeat protocol lets clients signal they are still active.

---

## What It Does

- **User Profiles** — Store and retrieve user metadata (name, avatar, bio, timezone, status, language)
- **Partial Updates** — Update any profile field independently (PATCH semantics)
- **List & Search** — Retrieve users by ID or search by name/email
- **Presence Tracking** — Track online/offline state via heartbeat signals
- **Presence Stream** — Real-time SSE stream of presence events (online/offline)
- **Local JWT Validation** — Verify access tokens using the RS256 public key (no auth service call)

---

## HTTP API

All endpoints return JSON. Timestamps are Unix seconds (UTC).

### Public Endpoints

#### POST `/api/v1/users/`

Create a user profile. Called by auth service on register (gRPC), but also available via HTTP.

**Request Body:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com"
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "name": "",
  "avatar": "",
  "bio": "",
  "timezone": "UTC",
  "status": "active",
  "language": "en",
  "created_at": 1639747350,
  "updated_at": 1639747350
}
```

**Error Responses:**
- `409 Conflict` — User already exists

---

#### GET `/api/v1/users/:id`

Get a user profile by UUID (public endpoint; no auth required).

**Response (200 OK):** Same as POST response.

**Error Responses:**
- `404 Not Found` — User not found

---

### Protected Endpoints (Bearer Token)

#### GET `/api/v1/users`

List all users with pagination.

**Query Parameters:**
```
GET /api/v1/users?limit=20&offset=0
```

| Param | Type | Default | Max | Description |
|-------|------|---------|-----|-------------|
| `limit` | int | 20 | 100 | Number of users per page |
| `offset` | int | 0 | — | Offset from start (0-indexed) |

**Response (200 OK):**
```json
{
  "users": [
    { "id": "...", "email": "...", "name": "...", ... },
    ...
  ],
  "total": 150
}
```

---

#### GET `/api/v1/users/search`

Search users by name or email (substring match, case-insensitive).

**Query Parameters:**
```
GET /api/v1/users/search?q=alice&limit=10&offset=0
```

| Param | Type | Max | Description |
|-------|------|-----|-------------|
| `q` | string | 100 chars | Search query (ILIKE on name or email) |
| `limit` | int | 100 | Results per page |
| `offset` | int | — | Pagination offset |

**Response (200 OK):** Same as `/users`.

**Note:** Special characters (`\`, `%`, `_`) are escaped for safe LIKE matching.

---

#### GET `/api/v1/users/me`

Get the authenticated user's own profile.

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):** Same as POST response.

---

#### PATCH `/api/v1/users/me`

Update the authenticated user's profile (partial update; omitted fields are unchanged).

**Request Body:**
```json
{
  "name": "Alice Smith",
  "avatar": "https://example.com/avatar.jpg",
  "bio": "Software engineer, coffee enthusiast",
  "timezone": "Europe/Lisbon",
  "status": "busy",
  "language": "pt"
}
```

**Field Constraints:**
- `name` — max 100 characters
- `avatar` — max 2048 characters (URL)
- `bio` — max 500 characters
- `timezone` — max 64 characters (IANA timezone ID, e.g., `"Europe/Lisbon"`)
- `status` — max 50 characters (e.g., `"active"`, `"busy"`, `"away"`)
- `language` — max 10 characters; **must be `"en"` or `"pt"` (allowlist)**

**Response (200 OK):** Updated profile object.

**Error Responses:**
- `400 Bad Request` — Invalid language (not `"en"` or `"pt"`)
- `401 Unauthorized` — Invalid token

---

#### POST `/api/v1/users/heartbeat`

Signal presence (user is online). Call this every 10 seconds from the client.

**Response (204 No Content):** Empty body.

**Side Effect:** Updates `last_seen_at` in the DB; if the user was marked offline, broadcasts an online event to all SSE subscribers.

---

#### GET `/api/v1/users/presence/stream`

Open an SSE stream to receive presence events (online/offline).

**Query Parameters:**
```
GET /api/v1/users/presence/stream?token=<access_token>
```

| Param | Description |
|-------|-------------|
| `token` | JWT access token (passed as query param because SSE/EventSource cannot set custom headers) |

**Stream Events:**

```
event: online
data: {"user_id": "550e8400-e29b-41d4-a716-446655440000"}

event: offline
data: {"user_id": "550e8400-e29b-41d4-a716-446655440000"}

: heartbeat
```

**Behavior:**
- On connect: sends a snapshot of all currently-online users
- Then: broadcasts presence changes in real-time
- Every 15 seconds: sends a keepalive comment (`: heartbeat`)
- On client disconnect: stream closes gracefully

**Error Responses:**
- `401 Unauthorized` — Invalid token

---

#### GET `/__health`

Health check endpoint (always returns 200, no auth required).

---

## gRPC API

The gRPC service runs on port 50052. RPCs are defined in `proto/users.proto`.

| RPC | Request | Response |
|-----|---------|----------|
| `CreateUser` | `id` (UUID string), `email` | `UserResponse` (user profile) |
| `GetUser` | `id` (UUID string) | `UserResponse` |
| `BatchGetUsers` | `ids` (repeated UUID strings) | `BatchGetUsersResponse` (`users[]`) |
| `UpdateProfile` | `id` + optional fields (`name`, `avatar`, `bio`, `timezone`, `status`, `language`) | `UserResponse` |
| `ListUsers` | `limit`, `offset` | `ListUsersResponse` (`users[]`, `total`) |
| `SearchUsers` | `query`, `limit`, `offset` | `SearchUsersResponse` (`users[]`, `total`) |

---

## Presence System

Presence is a three-component system:

### 1. Client Heartbeat

Clients POST `/api/v1/users/heartbeat` every 10 seconds (or whenever they want to signal "I'm still here").

### 2. Presence Tracker (In-Memory)

The `PresenceTracker` maintains:
- A map of `userId → last_heartbeat_time`
- A map of `userId → online_status` (boolean)

On heartbeat:
- If user transitions from offline → online: update DB (`last_seen_at = NOW()`), broadcast online event
- Otherwise: just update the last-seen time (no DB write)

### 3. Presence Reaper (Background Job)

Every 10 seconds, the reaper checks for users with no heartbeat for >25 seconds:
- Mark user offline in memory
- Update DB (`last_seen_at = NOW()`)
- Broadcast offline event to all SSE subscribers

```
Client heartbeat (every 10s)
     ↓
Tracker.Heartbeat(userId)
     ↓ [offline→online transition?]
   [YES] DB: UPDATE last_seen_at; Broker.Publish("online")
   [NO]  Update in-memory timestamp

Reaper tick (every 10s)
     ↓
  For each user with no heartbeat >25s:
     ↓
  Mark offline; DB UPDATE last_seen_at; Broker.Publish("offline")
```

### SSE Stream Behavior

On `/presence/stream` connection:
1. Send snapshot: `event: online data: {user_id: "..."}` for all currently-online users
2. Subscribe to the in-memory broker
3. Forward all online/offline events in real-time
4. Every 15 seconds, send a keepalive comment (`: heartbeat`)

Slow/disconnected subscribers that can't keep up with events are silently dropped (non-blocking publish).

---

## Partial Update Strategy

The `PATCH /api/v1/users/me` endpoint implements **merge-based** partial updates:

1. Read the current profile from the DB
2. Apply the non-null request fields to the in-memory object (overwrite)
3. Write all fields back to the DB

This pattern is necessary because the database layer (sqlc) generates plain `string` parameters (not nullable pointers), so we can't distinguish between "not provided" and "explicitly set to empty string" at the DB layer.

---

## Language Allowlist

The `language` field only accepts:
- `"en"` — English
- `"pt"` — Portuguese

Any other value returns a `400 Bad Request` error. This enforces consistency with the frontend's i18n setup.

---

## Database Schema

### `user_profiles` Table

```sql
CREATE TABLE user_profiles (
    id         UUID PRIMARY KEY,
    email      TEXT UNIQUE NOT NULL,         -- max 254 chars
    name       TEXT NOT NULL DEFAULT '',     -- max 100 chars
    avatar     TEXT NOT NULL DEFAULT '',     -- max 2048 chars
    bio        TEXT NOT NULL DEFAULT '',     -- max 500 chars
    timezone   TEXT NOT NULL DEFAULT 'UTC',  -- max 64 chars
    status     TEXT NOT NULL DEFAULT 'active', -- max 50 chars
    language   TEXT NOT NULL DEFAULT 'en',   -- max 10 chars ('en' or 'pt')
    last_seen_at TIMESTAMP,                  -- nullable; updated by heartbeat/reaper
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_email ON user_profiles(email);
CREATE INDEX idx_user_profiles_status ON user_profiles(status);
```

**Notes:**
- `language` was added in migration 0003
- `last_seen_at` was added in migration 0002 (for presence tracking)
- All text columns have `NOT NULL DEFAULT ''` to simplify queries (no nullable columns)

---

## Environment Variables

All required at startup.

| Variable | Description | Example |
|----------|-------------|---------|
| `USERS_DATABASE_URL` | PostgreSQL DSN | `postgres://postgres:password@localhost:5435/users?sslmode=disable` |
| `JWT_PUBLIC_KEY` | RS256 public key (PEM format) | (multi-line PEM content) |
| `USERS_GRPC_PORT` | Port for gRPC server | `50052` |
| `USERS_HTTP_PORT` | Port for HTTP server | `8082` |

**Note:** The users service needs **only the public key** (for JWT verification). It never issues tokens and does not need the private key.

---

## JWT Validation (Local)

Token validation happens locally without calling the auth service:

1. Parse the JWT using the RS256 public key
2. Verify signature and expiry
3. Extract `user_id` from the `sub` claim

This is fast (no RPC call) but means **token revocation is not checked** at the users service. The messaging service and channels service also validate locally, so revoked tokens might be accepted by them for a few seconds until the session expiry is reached.

---

## Development

### Run Migrations

```bash
cd services/users
export DATABASE_URL="postgres://postgres:password@localhost:5435/users?sslmode=disable"
make migrate-up
```

### Run Service Locally

```bash
export USERS_DATABASE_URL="postgres://postgres:password@localhost:5435/users?sslmode=disable"
export JWT_PUBLIC_KEY="$(cat ../../dev/public.pem)"
export USERS_GRPC_PORT=50052
export USERS_HTTP_PORT=8082

make run
```

### Run Tests

```bash
cd services/users
make test
```

### Useful Make Targets

| Target | Description |
|--------|-------------|
| `generate` | Recompile proto files |
| `build` | Build binary |
| `run` | Build and run service |
| `test` | Run unit tests |
| `tidy` | `go mod tidy` |
| `migrate-up` | Apply pending migrations |
| `migrate-down` | Rollback one migration |
| `install-tools` | Install build tools (including sqlc) |

---

## Error Handling & Status Codes

| Error | HTTP Status | gRPC Code | Reason |
|-------|-------------|-----------|--------|
| `ErrUserNotFound` | 404 | `NotFound` | User profile not found |
| `ErrUserAlreadyExists` | 409 | `AlreadyExists` | Email already registered |
| `ErrInvalidLanguage` | 400 | `InvalidArgument` | Language not in allowlist (`"en"`, `"pt"`) |

---

## Timezone Recommendations

The `timezone` field accepts any string (no validation). For consistency, use IANA timezone IDs:
- `"UTC"`
- `"Europe/Lisbon"`
- `"Europe/London"`
- `"America/New_York"`
- etc.

See [IANA Time Zone Database](https://www.iana.org/assignments/timezone-identifiers/) for the full list.

---

## Presence Timeout Tuning

Current configuration:
- **Heartbeat interval** (client): 10 seconds
- **Reaper check interval** (server): 10 seconds
- **Timeout threshold**: 25 seconds (2.5 heartbeat intervals)

If you adjust heartbeat intervals, update the reaper threshold to `2.5× client_interval`.

---

## Inter-Service Dependencies

The users service has **no runtime dependencies** on other services. It:
- Validates JWTs locally (doesn't call auth)
- Is called by auth (on register) and channels (to validate membership)
- Does not call any other services
