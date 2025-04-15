#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Define the output binary name
BINARY_NAME="you-go-server"
BINARY_PATH="./bin/${BINARY_NAME}" # Place binary in ./bin/

# Define the main package path
MAIN_PACKAGE="youGo/cmd/api"

# (Optional) Load environment variables from .env.local if it exists
if [ -f .env.local ]; then
  echo "Loading environment variables from .env.local"
  export $(grep -v '^#' .env.local | xargs)
fi

# (Optional) Load environment variables from .env if it exists and .env.local doesn't
# elif [ -f .env ]; then
#   echo "Loading environment variables from .env"
#   export $(grep -v '^#' .env | xargs)
# fi

# Ensure the output directory exists
mkdir -p ./bin

echo "Building Go application..."
# Build the Go application
# -o: specify output path
# ./cmd/api: path to the main package
go build -o "${BINARY_PATH}" "${MAIN_PACKAGE}"

echo "Build complete: ${BINARY_PATH}"

echo "Starting application..."
# Run the compiled binary
# Pass any command-line arguments if your app uses them: ${BINARY_PATH} arg1 arg2
"${BINARY_PATH}"

# Optional: Use a live-reloading tool like 'air' for development
# Ensure air is installed: go install github.com/cosmtrek/air@latest
# Then comment out the go build/run lines above and uncomment the line below:
# echo "Starting application with air (live reload)..."
# air
