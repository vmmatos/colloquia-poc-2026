# Auth Service

**Stateful JWT authentication, session management, and account lockout enforcement.**

The Auth service is the security boundary of the Colloquia system. It issues RS256 JWT access tokens, manages opaque refresh tokens with rotation and revocation, validates credentials, and enforces account lockout policies. Tokens are stored as SHA-256 hashes, never as plaintext.

---

## What It Does

- **User Registration** — Create credentials (email + password), hash with bcrypt (cost 12)
- **Login & Logout** — Verify credentials, issue token pair; revoke sessions on logout
- **Token Refresh** — Issue new token pair from a valid refresh token; atomically revoke old session (rotation)
- **Token Validation** — Verify JWT signature and check revocation status against DB
- **Account Lockout** — After 5 failed login attempts, lock account for 15 minutes
- **JWKS Endpoint** — Expose RS256 public key for other services and the KrakenD gateway

---

## Authentication Design

### RS256 (Asymmetric Keys)

The auth service uses **RS256 (RSA-256)** JWT signing, which allows:
- Auth service signs tokens with the **private key** (kept secret)
- All other services verify tokens with the **public key** (shared freely)
- Gateway caches the public key at startup from the JWKS endpoint
- No service-to-service authentication call needed on every request

### Token Pair Strategy

| Token | Type | TTL | Storage | Use Case |
|-------|------|-----|---------|----------|
| **Access Token** | JWT (RS256) | 15 minutes | Client session | Every authenticated request to the gateway; passed in `Authorization: Bearer` header |
| **Refresh Token** | Opaque (random) | 7 days | HttpOnly cookie (via BFF) + DB hash | Obtain a new access token without re-entering credentials |

### Token Hashing

Tokens are **never stored in plaintext**:
- Access tokens: SHA-256 hash stored in `sessions.access_token_hash`
- Refresh tokens: SHA-256 hash stored in `sessions.refresh_token_hash` (unique constraint)

This prevents token leaks from the auth DB from compromising user accounts.

### Stateful Validation

Token validation is **stateful** (not just JWT signature verification):

1. Verify JWT signature and expiry (standard JWT checks)
2. Query the DB for the session row matching the hashed token
3. Check that `revoked = false`
4. If any step fails, reject the token

This enables **true logout**: revoking a token revokes its session row, and future requests with that token will fail even if the JWT signature is still valid.

### Token Rotation

On `RefreshToken` call:
1. Verify the old refresh token hash matches a non-revoked session
2. Create a new session with a new token pair
3. Set the old session to `revoked = true` atomically
4. Return the new tokens

This prevents token replay: if an attacker steals a refresh token and uses it, legitimate refresh attempts after that will fail because the session is now revoked.

### Account Lockout

After 5 failed login attempts:
- Set `users.locked_until = NOW() + 15 minutes`
- Reject all further login attempts until the lockout expires
- Reset `failed_login_attempts` to 0 on successful login

---

## HTTP API

All endpoints return JSON responses.

### Public Endpoints (No Authentication Required)

#### POST `/api/v1/auth/register`

Create a new user account.

**Request Body:**
```json
{
  "email": "alice@example.com",
  "password": "SecurePassword123"
}
```

**Constraints:**
- `email`: max 254 characters
- `password`: 8–128 characters

**Response (201 Created):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImNvbGxvcXVpYS1hdXRoLWtleS0xIn0.eyJ1c2VyX2lkIjoiNTUwZTg0MzAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwiaWF0IjoxNjM5NzQ3MzUwLCJleHAiOjE2Mzk3NDc5NTB9.sig...",
  "refresh_token": "rPkK7mL0zXqJ_dGtN1bHcA",
  "expires_at": "2024-12-17T12:15:50Z"
}
```

**Error Responses:**
- `409 Conflict` — Email already registered (`ErrEmailAlreadyExists`)
- `400 Bad Request` — Validation error (email or password invalid)

**Side Effect:** Auth service calls the **users service gRPC** (`CreateUser`) to create a profile. If users service is unavailable, registration succeeds anyway (best-effort).

---

#### POST `/api/v1/auth/login`

Authenticate a user with email and password.

**Request Body:**
```json
{
  "email": "alice@example.com",
  "password": "SecurePassword123"
}
```

**Response (200 OK):** Same as `/register`.

**Error Responses:**
- `401 Unauthorized` — Invalid credentials (`ErrInvalidCredentials`)
- `403 Forbidden` — Account locked due to too many failed attempts (`ErrAccountLocked`); retry after 15 minutes

---

#### POST `/api/v1/auth/refresh`

Obtain a new access token using a refresh token.

**Request Body:**
```json
{
  "refresh_token": "rPkK7mL0zXqJ_dGtN1bHcA"
}
```

**Response (200 OK):** Same as `/register`.

**Error Responses:**
- `404 Not Found` — Refresh token not found or revoked (`ErrSessionNotFound`)
- `401 Unauthorized` — Token expired (`ErrTokenExpired`)

---

#### GET `/.well-known/jwks.json`

Expose the RS256 public key in JWKS (JSON Web Key Set) format, used by KrakenD gateway and other services for token verification.

**Response (200 OK):**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "colloquia-auth-key-1",
      "n": "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtV...",
      "e": "AQAB"
    }
  ]
}
```

**Cache:** Clients SHOULD cache this for up to 1 hour.

---

### Protected Endpoints (Bearer Token Required)

#### POST `/api/v1/auth/logout`

Revoke the current session (access token).

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Response (204 No Content):** Empty body.

**Error Responses:**
- `401 Unauthorized` — Invalid or expired token

---

#### GET `/api/v1/auth/validate`

Check if an access token is valid. **Always returns 200** (never 401), for use by clients checking token validity before making requests.

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "valid": true,
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Invalid Token Response (200 OK):**
```json
{
  "valid": false,
  "user_id": null
}
```

---

#### GET `/api/v1/auth/me`

Get the authenticated user's email and ID.

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "valid": true,
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com"
}
```

---

#### GET `/__health`

Health check endpoint (always returns 200, no auth required).

---

## gRPC API

The gRPC service runs on port 50051. RPCs are defined in `proto/auth.proto`.

### Unary RPCs

| RPC | Request | Response |
|-----|---------|----------|
| `Register` | `email`, `password` | `AuthResponse` (`user_id`, `access_token`, `refresh_token`, `expires_at`) |
| `Login` | `email`, `password` | `AuthResponse` |
| `Logout` | `access_token` | `LogoutResponse` (empty) |
| `RefreshToken` | `refresh_token` | `AuthResponse` |
| `ValidateToken` | `access_token` | `ValidateTokenResponse` (`valid`, `user_id`, `email`) |

---

## Database Schema

The auth service uses two tables:

### `users` Table

Stores user credentials.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,           -- max 254 chars
    password_hash TEXT NOT NULL,          -- bcrypt hash, max 512 chars
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP NULL,          -- NULL = not locked
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Constraints:**
- `email` unique (no duplicate accounts)
- `password_hash` max 512 (bcrypt hashes are ~60 chars, but allow headroom)

### `sessions` Table

Stores active access and refresh token hashes.

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT UNIQUE NOT NULL,
    access_token_hash TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_access_token_hash ON sessions(access_token_hash);
CREATE INDEX idx_sessions_refresh_token_hash ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
```

**Flow:**
1. `Login` → create `users` row, create `sessions` row with hashed tokens
2. `RefreshToken` → verify old session not revoked, create new session, set old to `revoked = true`
3. `ValidateToken` → hash incoming token, query `sessions` table, verify not revoked
4. `Logout` → set `revoked = true` on the session row

---

## Environment Variables

All required at startup. Service exits with an error if any are missing.

| Variable | Description | Example |
|----------|-------------|---------|
| `AUTH_DATABASE_URL` | PostgreSQL DSN | `postgres://postgres:password@localhost:5434/auth?sslmode=disable` |
| `JWT_PRIVATE_KEY` | RS256 private key (PEM format) | (multi-line PEM content) |
| `JWT_PUBLIC_KEY` | RS256 public key (PEM format) | (multi-line PEM content) |
| `AUTH_GRPC_PORT` | Port for gRPC server | `50051` |
| `AUTH_HTTP_PORT` | Port for HTTP server | `8081` |
| `USERS_GRPC_ADDRESS` | Address of users service (best-effort) | `localhost:50052` |

**Key Format:** PEM-encoded RSA keys can be raw multi-line strings or base64-encoded single-line strings; the service auto-detects and decodes both formats.

---

## Development

### Generate RSA Keys (One-Time)

```bash
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -pubout -out public.pem
```

### Run Migrations

```bash
cd services/auth
export DATABASE_URL="postgres://postgres:password@localhost:5434/auth?sslmode=disable"
make migrate-up
```

### Run Service Locally

```bash
export AUTH_DATABASE_URL="postgres://postgres:password@localhost:5434/auth?sslmode=disable"
export JWT_PRIVATE_KEY="$(cat private.pem)"
export JWT_PUBLIC_KEY="$(cat public.pem)"
export USERS_GRPC_ADDRESS="localhost:50052"
export AUTH_GRPC_PORT=50051
export AUTH_HTTP_PORT=8081

make run
```

### Run Tests

```bash
cd services/auth
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
| `install-tools` | Install build tools |

---

## Inter-Service Dependencies

On `Register`, the auth service calls the **users service** gRPC `CreateUser` RPC to create a user profile. This is a **best-effort** call: if the users service is unavailable, the error is logged and registration proceeds. The user's auth credentials are persisted even if the profile creation fails.

---

## Error Handling & Status Codes

| Error | HTTP Status | gRPC Code | Reason |
|-------|-------------|-----------|--------|
| `ErrEmailAlreadyExists` | 409 | `AlreadyExists` | Email already registered |
| `ErrInvalidCredentials` | 401 | `Unauthenticated` | Wrong email or password |
| `ErrAccountLocked` | 403 | `PermissionDenied` | Account locked (5 failed attempts) |
| `ErrSessionNotFound` | 404 | `NotFound` | Refresh token not found or revoked |
| `ErrTokenExpired` | 401 | `Unauthenticated` | Access or refresh token expired |
| `ErrTokenInvalid` | 401 | `Unauthenticated` | Invalid token format or signature |

---

## Bcrypt Configuration

- **Cost**: 12 (balanced security vs performance; ~100ms per hash on modern hardware)
- **Algorithm**: bcrypt with Blowfish cipher
- Password hashing happens in-memory; hashes are never logged

---

## Session Cleanup

Old revoked sessions are **not automatically deleted** from the database. Consider adding a background job to delete sessions older than 30 days, or use database expiration policies (PostgreSQL `timescaledb` or custom VACUUM).

---

## Inter-Service Communication

**Best-Effort Call (Non-Blocking):**
- Auth → Users (on register)
- If unavailable, registration still succeeds

**Fail-Closed Calls:**
- None; auth has no hard dependencies on other services
