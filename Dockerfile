# ── UI build stage ──────────────────────────────────────────────────────────────
FROM node:22-alpine AS ui-builder

WORKDIR /src

COPY cmd/archipulse/ui/package.json cmd/archipulse/ui/package-lock.json ./cmd/archipulse/ui/
RUN npm ci --prefix cmd/archipulse/ui

COPY cmd/archipulse/ui/ ./cmd/archipulse/ui/
RUN npm run build --prefix cmd/archipulse/ui

# ── Go build stage ───────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=ui-builder /src/cmd/archipulse/ui/dist ./cmd/archipulse/ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /archipulse ./cmd/archipulse

# ── Runtime stage ───────────────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /archipulse ./archipulse
COPY migrations/ ./migrations/
COPY entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
