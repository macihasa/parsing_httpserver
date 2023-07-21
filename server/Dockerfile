# syntax=docker/dockerfile:1

FROM golang:1.19

# Set the working directory to the app directory
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the source code and pkg folder to the working directory
COPY / ./


# Build the Go application
RUN go build -o main ./src

# Expose the desired port (change if needed)
EXPOSE 5000

# Run the Go application
CMD ["./main"]
