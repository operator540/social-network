# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o social-network ./cmd/api

# Финальный образ
FROM alpine:3.20

RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/social-network .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./social-network"]
