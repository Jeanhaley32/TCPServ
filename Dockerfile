# Docker config file for tcp serv
# run with ```docker run -p 6000:6000 -e PORT=6000 tcpserv`
# Builder stage (pulls from GitHub)
FROM golang:alpine AS builder
ENV PORT=6000

RUN apk add --no-cache git

RUN git clone -b main https://github.com/jeanhaley32/tcpserv /app

WORKDIR /app/

# Install go dependencies
RUN go mod download

# Build the Go binary
RUN go build -o tcpserv .

# Runtime stage (uses pre-built binary)
FROM alpine:latest AS runtime

# Copy the built binary from the builder stage
COPY --from=builder /app/tcpserv /app/msg.txt /app/

# Expose the port
EXPOSE $PORT

WORKDIR /app/

# Run App with port flag
CMD ["sh", "-c", "/app/tcpserv --port=${PORT}"]
