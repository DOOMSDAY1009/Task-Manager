# --- build stage ---
FROM golang:1.25-alpine AS build
WORKDIR /app

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build a static binary so we can run it on a tiny base image.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# --- run stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
COPY --from=build /server /server
USER appuser
EXPOSE 8080
ENTRYPOINT ["/server"]
