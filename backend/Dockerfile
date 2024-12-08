# Use official Golang image to build the backend
FROM golang:1.21 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules files
COPY backend/go.mod backend/go.sum ./

# Download Go modules dependencies
RUN go mod tidy

# Copy the entire backend code
COPY backend .

# Build the Go application
RUN go build -o main ./app/main.go

# Create a minimal image to run the Go application
FROM alpine:latest

# Install required libraries
RUN apk add --no-cache ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the Go application from the builder stage
COPY --from=builder /app/main .

# Expose the port your app will run on
EXPOSE 8080

# Command to run the Go application
CMD ["./main"]
