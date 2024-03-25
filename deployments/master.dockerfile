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

# Navigate to the master directory.
WORKDIR /app/cmd/master

# Build the master component of the application.
RUN go build -o master

FROM rabbitmq:3.13.0-management

COPY --from=builder /app/cmd/master/master /app/master

# Copy necessary resources from your project.
COPY --from=builder /app/config /app/config
COPY --from=builder /app/pkg /app/pkg

COPY master-start.sh /app/start.sh
RUN chmod +x /app/start.sh

# Set the working directory
WORKDIR /app

# Expose the port the application uses.
EXPOSE 8080 15672 5671 5672

CMD ["/app/start.sh"]
