#!/bin/bash

PORT=6969

# Find process IDs using the port
PIDS=$(sudo lsof -ti :$PORT)

if [ -z "$PIDS" ]; then
    echo "No processes found using port $PORT"
else
    echo "Killing processes using port $PORT: $PIDS"
    sudo kill -9 $PIDS
    echo "Processes killed."
fi
