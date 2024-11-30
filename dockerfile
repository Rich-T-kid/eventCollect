# Use the Go base image with Alpine Linux
FROM golang:1.21-alpine

# Install build dependencies for CGO and SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Enable CGO for SQLite support and tidy dependencies
ENV CGO_ENABLED=1
RUN go mod tidy

# Copy all source code into the container
COPY . .

# Copy the .env file for application settings (if needed)
COPY .env .env

# Build the Go application
RUN go build -o main .

# Set the entrypoint to run the built Go application
CMD ["./main"]
