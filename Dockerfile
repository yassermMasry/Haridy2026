FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /haridy-erp ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -H appuser
COPY --from=builder /haridy-erp /usr/local/bin/haridy-erp
COPY templates ./templates
COPY static ./static
COPY migrations ./migrations
USER appuser
EXPOSE 8080
CMD ["haridy-erp"]
