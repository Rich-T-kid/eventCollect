# Use the Go base image with Alpine Linux
FROM golang:1.21-alpine

# Install build dependencies for CGO and SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency installation
COPY go.mod go.sum ./

# Enable CGO and download dependencies
ENV CGO_ENABLED=1
RUN go mod tidy

# Copy the application source code
COPY . .

# Copy the .env file
COPY .env .env


# Build the Go application
RUN go build -o main .

# Run the built application
CMD ["./main"]
