#!/bin/sh

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# Load .env file if it exists in the parent directory (for local execution)
if [ -f ".env" ]; then
  # Use 'export' with 'source' or '.' to make variables available to migrate command
  echo "Loading environment variables from .env file"
  export $(grep -v '^#' .env | xargs)
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL environment variable is not set."
  exit 1
fi

MIGRATIONS_DIR="./migrations" # Path to your migrations folder

# --- Command Logic ---
COMMAND=$1

# Print usage instructions if no command is given
if [ -z "$COMMAND" ]; then
  echo "Usage: $0 <command>"
  echo "Commands:"
  echo "  up          Apply all available migrations"
  echo "  down [N]    Revert N last migrations (default: 1)"
  echo "  force <V>   Set migration version V forcefully (use with caution)"
  echo "  version     Print current migration version"
  exit 1
fi

# Execute migrate command
case "$COMMAND" in
  up)
    echo "Applying migrations..."
    migrate -database "$DATABASE_URL" -path "$MIGRATIONS_DIR" up
    ;;
  down)
    COUNT=${2:-1} # Default to reverting 1 migration if no number is provided
    echo "Reverting last $COUNT migration(s)..."
    migrate -database "$DATABASE_URL" -path "$MIGRATIONS_DIR" down "$COUNT"
    ;;
  force)
    VERSION=$2
    if [ -z "$VERSION" ]; then
      echo "Error: Please specify a version number for force."
      exit 1
    fi
    echo "Forcing migration version to $VERSION..."
    migrate -database "$DATABASE_URL" -path "$MIGRATIONS_DIR" force "$VERSION"
    ;;
  version)
    echo "Checking migration version..."
    migrate -database "$DATABASE_URL" -path "$MIGRATIONS_DIR" version
    ;;
  *)
    echo "Error: Unknown command '$COMMAND'"
    exit 1
    ;;
esac

echo "Migration command '$COMMAND' executed successfully."