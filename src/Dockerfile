# ---------- Stage 1: build ----------
FROM golang:1.26.4 AS builder
WORKDIR /app

# Copy dependency files first so this layer is cached
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o product-service .
# ---------- Stage 2: run ----------
FROM --platform=linux/amd64 alpine:3.20
WORKDIR /app
COPY --from=builder /app/product-service .

ENV GIN_MODE=release
EXPOSE 8080
CMD ["./product-service"]