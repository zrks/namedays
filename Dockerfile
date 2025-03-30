# Use the official Golang image as the base image
FROM golang:1.20-alpine AS builder

RUN apk add --no-cache gcc musl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

ENV CGO_ENABLED=1

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN GOMEMLIMIT=400MiB go build -o main .

# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/
COPY db-ops/namedays.json ./db-ops/namedays.json

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
