FROM golang:1.19 AS builder

WORKDIR /app
COPY . .

# Download dependencies and build the app.
RUN go mod download
RUN CGO_ENABLED=0 go build -o pubsub .

# ============================================================================

FROM scratch

# Copy the binary from the builder.
COPY --from=builder /app/pubsub .

CMD ["./pubsub"]
