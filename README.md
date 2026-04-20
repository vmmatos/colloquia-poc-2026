# Colloquia POC 2026

**Slack-like team messaging platform built as a Go microservices proof-of-concept.**

Colloquia is an exploration of distributed systems patterns through a real-world team communication platform. It features real-time messaging in public channels and direct messages, user presence tracking, and AI-powered reply suggestions integrated with a local LLM. The entire system runs on a single machine via Docker Compose, making it ideal for learning microservices architecture, gRPC, database-per-service patterns, and event-driven real-time communication.

---

## System Architecture

```mermaid
graph LR
    Browser["🌐 Browser"]
    Ingress["NGINX Ingress<br/>Port 80"]
    Web["Nuxt 4 SPA<br/>Port 3000"]
    Gateway["KrakenD Gateway<br/>Port 8000<br/>JWT Validation • Rate Limits • CORS"]
    
    Auth["Auth Service<br/>Port 8081/50051"]
    Users["Users Service<br/>Port 8082/50052<br/>Presence • Profiles"]
    Channels["Channels Service<br/>Port 8083/50053<br/>Channels • DMs • Members"]
    Messaging["Messaging Service<br/>Port 8084/50054<br/>Messages • SSE Stream"]
    Assist["Assist Service<br/>Port 8085<br/>AI Suggestions"]
    
    Ollama["Ollama LLM<br/>Port 11434"]
    
    AuthDB["PostgreSQL<br/>auth"]
    UsersDB["PostgreSQL<br/>users"]
    ChannelsDB["PostgreSQL<br/>channels"]
    MessagingDB["PostgreSQL<br/>messaging"]
    
    Browser -->|"/ (SPA)"| Ingress
    Browser -->|"/api/** (API)"| Ingress
    Ingress -->|"Port 3000"| Web
    Ingress -->|"Port 8000"| Gateway
    
    Web -->|"HTTP"| Gateway
    
    Gateway -->|"Port 8081"| Auth
    Gateway -->|"Port 8082"| Users
    Gateway -->|"Port 8083"| Channels
    Gateway -->|"Port 8084"| Messaging
    Gateway -->|"Port 8085"| Assist
    
    Auth -->|"gRPC CreateUser<br/>Port 50052"| Users
    Channels -->|"gRPC ValidateUser<br/>Port 50052"| Users
    Messaging -->|"gRPC ValidateMembership<br/>Port 50053"| Channels
    Assist -->|"gRPC GetMessages<br/>Port 50054"| Messaging
    Assist -->|"HTTP /api/generate<br/>Port 11434"| Ollama
    
    Auth --> AuthDB
    Users --> UsersDB
    Channels --> ChannelsDB
    Messaging --> MessagingDB
    
    style Gateway fill:#f9a825,color:#000
    style Ollama fill:#437eba,color:#fff
    style Ingress fill:#009639,color:#fff
```

---

## Services Overview

| Service | Purpose | HTTP Port | gRPC Port | DB |
|---------|---------|-----------|-----------|-----|
| **auth** | JWT token issuance, validation, refresh; session management; account lockout | 8081 | 50051 | ✓ |
| **users** | User profiles, presence tracking, list/search | 8082 | 50052 | ✓ |
| **channels** | Channels, DMs, groups, membership management | 8083 | 50053 | ✓ |
| **messaging** | Message storage, real-time fan-out via SSE | 8084 | 50054 | ✓ |
| **assist** | AI-powered reply suggestions via Ollama | 8085 | — | — |

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Nuxt 4 (Vue 3), Tailwind CSS, i18n (EN/PT) |
| **Backend** | Go 1.23+, gRPC, Protocol Buffers |
| **Databases** | PostgreSQL 16 (one per service) |
| **Gateway** | KrakenD v3 (JWT validation, rate limiting, CORS) |
| **Ingress** | NGINX |
| **LLM** | Ollama (local, no cloud dependency) |
| **IPC** | gRPC + HTTP |
| **Real-Time** | Server-Sent Events (SSE) |

---

## Key Design Decisions

### Authentication & Authorization
- **RS256 JWT** — asymmetric keys; only auth service has private key; other services verify with shared public key
- **Token Strategy**: 
  - Access token (JWT): 15 minutes, issued on login/refresh
  - Refresh token (opaque): 7 days, stored as SHA-256 hash in auth DB
- **Stateful Validation**: Token revocation checked against DB, not just JWT signature — enables true logout
- **Token Rotation**: Refresh creates new session, revokes old one atomically
- **Account Lockout**: 5 failed login attempts trigger 15-minute lockout

### Data Architecture
- **Database-per-Service**: Each microservice owns its PostgreSQL schema (auth, users, channels, messaging)
- **No shared tables or transactions** across services
- **Eventual consistency** between services (no distributed transactions)

### Gateway Pattern
- **KrakenD** validates JWT signatures, enforces rate limits, handles CORS
- Services trust the `User-ID` header set by the gateway (already validated)
- Keeps auth logic at the perimeter; services are auth-agnostic

### Presence & Real-Time
- **SSE (Server-Sent Events)** for real-time messaging and presence updates
- **Heartbeat protocol**: clients POST `/heartbeat` every 10 seconds
- **Presence timeout**: users marked offline after 25 seconds without heartbeat
- **In-memory state**: presence is not persisted (session-scoped)

### AI Integration
- **Ollama**: local LLM runner, no external API calls
- **Caching**: suggestion results cached for 30 seconds per input (in-memory)
- **Graceful degradation**: LLM unavailability doesn't block messaging

### Inter-Service Communication
- **HTTP** for external (gateway → services)
- **gRPC** for internal service-to-service calls (low latency, type-safe)
- **Best-effort** calls are non-blocking (e.g., auth → users; if users is down, registration succeeds anyway)
- **Fail-closed** calls block the request (e.g., messaging → channels; if channels is down, message send fails)

---

## Repository Layout

```
colloquia-poc-2026/
├── README.md                 # This file
├── LICENSE                   # MIT
├── CONTRIBUTING.md
│
├── apps/
│   └── web/                  # Nuxt 4 SPA frontend
│       ├── app/              # Vue components, pages, composables
│       ├── server/api/       # BFF proxy routes (Nuxt)
│       ├── shared/           # Shared types
│       ├── i18n/             # i18n locale files (EN/PT)
│       ├── nuxt.config.ts
│       ├── package.json
│       └── README.md         # Frontend documentation
│
├── services/                 # Go microservices
│   ├── auth/                 # Authentication service
│   │   ├── cmd/api/
│   │   ├── internal/
│   │   ├── migrations/
│   │   ├── proto/
│   │   ├── Dockerfile
│   │   ├── Makefile
│   │   └── README.md
│   │
│   ├── users/                # User profiles & presence
│   │   ├── ... (same structure as auth)
│   │   └── README.md
│   │
│   ├── channels/             # Channels & membership
│   │   ├── ... (same structure)
│   │   └── README.md
│   │
│   ├── messaging/            # Messages & SSE
│   │   ├── ... (same structure)
│   │   └── README.md
│   │
│   └── assist/               # AI suggestions
│       ├── ... (same structure, no DB)
│       └── README.md
│
├── gateway/                  # KrakenD API gateway
│   ├── krakend.tmpl          # Gateway config template
│   ├── partials/             # Per-service route definitions
│   ├── jwks.json             # RS256 public key for JWT validation
│   └── Dockerfile
│
├── dev/                      # Local development
│   ├── docker-compose.yaml   # Full stack orchestration
│   ├── nginx.conf            # NGINX ingress config
│   ├── .dev.env              # Non-secret env vars
│   └── .secrets.env          # Secret env vars (not in git)
│
└── shared/                   # Placeholder for future shared code
    └── pkg/
```

---

## Quick Start (Docker Compose)

### Prerequisites
- **Docker Desktop** (with Docker Compose)
- **OpenSSL** (generate RSA key pair)
- **Port availability**: 80, 3000, 8000, 8001, 11434, 5434–5437

### 1. Generate RSA Keys (One-Time)

From the root directory:

```bash
openssl genrsa -out dev/private.pem 4096
openssl rsa -in dev/private.pem -pubout -out dev/public.pem
```

### 2. Configure Environment

Create `dev/.secrets.env` (not in git):

```bash
# JWT keys (generated above, multi-line PEM format is OK)
JWT_PRIVATE_KEY=$(cat dev/private.pem)
JWT_PUBLIC_KEY=$(cat dev/public.pem)

# Database URLs (docker-compose service names are hostnames inside the network)
AUTH_DATABASE_URL=postgres://postgres:password@auth-db:5432/auth?sslmode=disable
USERS_DATABASE_URL=postgres://postgres:password@users-db:5432/users?sslmode=disable
CHANNELS_DATABASE_URL=postgres://postgres:password@channels-db:5432/channels?sslmode=disable
MESSAGING_DATABASE_URL=postgres://postgres:password@messaging-db:5432/messaging?sslmode=disable

# Inter-service gRPC addresses (use docker-compose service names)
USERS_GRPC_ADDRESS=users-service:50052
CHANNELS_GRPC_ADDRESS=channels-service:50053
MESSAGING_GRPC_ADDRESS=messaging-service:50054

# Ollama
ASSIST_OLLAMA_BASE_URL=http://ollama:11434
ASSIST_OLLAMA_MODEL=qwen2.5:0.5b
ASSIST_OLLAMA_TIMEOUT_SECONDS=60
```

### 3. Start the Stack

```bash
cd dev
docker compose up
```

Services will be available at:
- **Frontend**: http://localhost (NGINX ingress)
- **SPA**: http://localhost:3000 (direct)
- **API Gateway**: http://localhost:8000
- **Individual services**: See Services Overview table above

Default credentials:
- **PostgreSQL** all databases: `postgres` / `password`
- **Ollama**: http://ollama:11434 (no auth)

### 4. First Run

The `docker compose` will:
1. Create databases and run migrations automatically
2. Pull and initialize the Ollama model (`qwen2.5:0.5b`) — this may take 1–2 minutes on first run
3. Pre-warm the LLM model (optional, via assist service startup hook)

Open http://localhost and register a new account.

---

## Local Development (Per-Service)

To run an individual service locally (e.g., for debugging), you need its dependencies running in Docker, but can run the service itself on your machine:

### Example: Run `auth` Service Locally

```bash
# Terminal 1: Start dependent services in Docker (users service, databases)
cd dev
docker compose up auth-db auth-migrate users-db users-migrate users-service

# Terminal 2: Run auth service locally
cd services/auth
export AUTH_DATABASE_URL="postgres://postgres:password@localhost:5434/auth?sslmode=disable"
export USERS_GRPC_ADDRESS="localhost:50052"  # users-service running in Docker
export JWT_PRIVATE_KEY=$(cat ../../dev/private.pem)
export JWT_PUBLIC_KEY=$(cat ../../dev/public.pem)
export AUTH_GRPC_PORT=50051
export AUTH_HTTP_PORT=8081

make run
```

(Adjust database ports and `GRPC_ADDRESS` values as needed for each service.)

---

## Architecture Highlights

### Microservices
Each service is independently deployable, scalable, and replaceable. Services communicate via:
- **gRPC** (internal, type-safe, low-latency)
- **HTTP + REST** (via gateway, external clients)

### Event-Driven Real-Time
Messages and presence updates are pushed to clients via **Server-Sent Events (SSE)**:
- Single connection per client for all subscribed channels
- Automatic reconnection with exponential backoff
- Gateway timeout is 0 (no timeout; SSE is long-lived)

### Database-per-Service
- No shared database
- Schema changes are scoped to one service
- Queries across services go through APIs (enforced consistency)
- Data duplication is intentional (eventual consistency)

### Stateless Services
- All state is either in the database or client session
- Services can be restarted without client impact
- Horizontal scaling is straightforward (add more instances behind a load balancer)

---

## Testing

Each service has unit tests:

```bash
cd services/auth
make test
```

The frontend has integration tests:

```bash
cd apps/web
yarn test
```

---

## Documentation

- **Backend Services**: See `services/{auth,users,channels,messaging,assist}/README.md`
- **Frontend**: See `apps/web/README.md`
- **Architecture Decisions**: See "Key Design Decisions" section above
- **API Reference**: Available in each service's README (HTTP endpoints + gRPC RPCs)

---

## Deployment Notes

This POC is built for **local development on a single machine**. For production deployment:

- Replace NGINX with a proper ingress controller (Kubernetes Ingress, AWS ALB, etc.)
- Use managed PostgreSQL (AWS RDS, Azure Database, etc.)
- Deploy services to Kubernetes or a container orchestration platform
- Replace local Ollama with a managed LLM API (OpenAI, Anthropic, etc.)
- Add CI/CD pipelines (GitHub Actions, GitLab CI, etc.)
- Add observability: logging aggregation (ELK, Datadog), tracing (Jaeger), metrics (Prometheus)
- Implement secrets management (HashiCorp Vault, AWS Secrets Manager, etc.)

---

## License

MIT License — See `LICENSE` file.

---

## Contributing

See `CONTRIBUTING.md` for guidelines on code style, commit messages, and pull requests.