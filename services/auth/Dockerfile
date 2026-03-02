FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth ./cmd/api/main.go

EXPOSE 50051

CMD ["./auth"]