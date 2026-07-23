# Stage 1: Build the Go binary using a compiler environment
FROM golang:alpine AS builder

WORKDIR /app

# Copy your go.mod and go.sum files first to leverage Docker's caching
COPY go.mod go.sum ./
RUN go mod download || true

# Copy the rest of your service's source code
COPY . .

# Compile the Go code into a single, highly optimized standalone binary named 'main'
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Create a tiny runtime environment
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Pull only the compiled binary out of Stage 1, leaving the heavy compiler behind
COPY --from=builder /app/main .

# Execute your app when the container starts
CMD ["./main"]