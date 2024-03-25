FROM golang:1.21.5 as builder

# Enable CGO for sqlite.
ENV CGO_ENABLED=1

# Install gcc for CGO dependencies.
RUN apt-get update && apt-get install -y gcc

# Set the working directory inside the container.
WORKDIR /app

# Copy go mod and sum files to leverage Docker cache.
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download

COPY . .

# Navigate to the worker directory.
WORKDIR /app/cmd/worker

# Build the worker component of the application.
RUN go build -o worker

FROM golang:1.21.5

COPY --from=builder /app/cmd/worker/worker /app/worker

# Copy necessary resources
COPY --from=builder /app/config /app/config
COPY --from=builder /app/pkg /app/pkg

# Set the working directory
WORKDIR /app

CMD ["./worker"]
