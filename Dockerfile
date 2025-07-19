# --- Этап 1: Сборка (Builder) ---
FROM golang:1.24.3-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/subtracker ./cmd/app

# --- Этап 2: Финальный образ (Final) ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/subtracker .
COPY --from=builder /app/docs ./docs
EXPOSE 8080
CMD ["./subtracker"]