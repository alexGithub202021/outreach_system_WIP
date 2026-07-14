# stage 1: create the go app executable binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependencies and source code
COPY go.mod ./
RUN go mod download
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

# Stage 2: Final lightweight image
FROM alpine:latest

# setup collection of valid certificates
RUN apk add --no-cache ca-certificates

# Copy the binary from the transient builder stage, that is discarded once the final image is built
COPY --from=builder /app/myapp /myapp

ENTRYPOINT ["/myapp"]