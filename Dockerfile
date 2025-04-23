# Build stage
FROM docker.io/golang:alpine AS builder

# Set working directory
WORKDIR /build_app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY ./rest ./rest
COPY ./rpc ./rpc
COPY ./cmd/horizon/index.html ./cmd/horizon/index.html
COPY ./cmd/horizon/main.go ./cmd/horizon/main.go
COPY ./*.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -gcflags "all=-trimpath=$GOPATH" -o horizon cmd/horizon/main.go

# Final stage
FROM scratch

WORKDIR /app

# Copy the built executable
COPY --from=builder /build_app/horizon ./

# Define the command to run the application
ENTRYPOINT ["./horizon"]