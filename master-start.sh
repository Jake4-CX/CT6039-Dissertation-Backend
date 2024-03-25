#!/bin/sh
# start.sh

# Start RabbitMQ in the background
echo "Starting RabbitMQ in the background..."
rabbitmq-server -detached

# Wait for RabbitMQ to fully start
echo "Waiting for RabbitMQ to be fully operational..."
rabbitmqctl wait /var/lib/rabbitmq/mnesia/rabbit@$(hostname).pid

# Start the master application
echo "Starting the master application..."
./master