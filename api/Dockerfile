# Dockerfile for the API
# Consists of a multi-stage build
# First it creates a go binary in a builder container
# Then it copies the binary to a smaller container
# The smaller container is the final image that is used to run the API

# Builder container
FROM golang:1.22 AS builder

# Set the working directory
WORKDIR /app

# Copy the go modules file
COPY go.mod .

# Copy the go sum file
COPY go.sum .

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Smaller container
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the builder container
COPY --from=builder /app/main .

# Expose the port
EXPOSE 8080

# Run the binary
CMD ["./main"]
