# colloquia-poc-2026
Where the teams talk about what they are doing and what they are planning to do.

## first plan

```text
colloquio/
│
├── apps/
│   ├── web/                # Vue 3
│   └── mobile/             # React Native
│
├── gateway/
│   └── krakend/
│       ├── krakend.json
│       └── config/
│
├── services/
│   ├── auth/
│   │   ├── cmd/
│   │   │   └── api/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── transport/
│   │   │   │   ├── http/
│   │   │   │   └── grpc/
│   │   │   └── config/
│   │   ├── proto/
│   │   ├── migrations/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── users/
│   ├── channels/
│   ├── messaging/
│   └── notifications/
│
├── shared/
│   ├── proto/              # shared contracts
│   ├── pkg/
│   │   ├── logger/
│   │   ├── middleware/
│   │   ├── jwt/
│   │   └── errors/
│
├── deploy/
│   ├── docker-compose.yml
│   └── k8s/
│
└── Makefile
```