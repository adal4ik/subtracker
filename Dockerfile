FROM golang:1.24.3-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY docs/ ./docs/

RUN go build -o subtracker ./cmd/app

CMD ["./subtracker"]
