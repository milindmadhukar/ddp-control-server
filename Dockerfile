# Start from the official Go image
FROM 1.22-alpine3.20

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port the app runs on
EXPOSE 8069

# Command to run the executable
CMD ["./main"]
