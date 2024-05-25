# Use the official Golang image as the base image
FROM golang:1.22.3-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code to the container
COPY . .

# Build the Go application
RUN go build -o main 

# Expose the port on which the API will run
EXPOSE 3001

# Command to run the executable
CMD ["./main"]
