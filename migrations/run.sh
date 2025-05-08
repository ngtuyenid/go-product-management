#!/bin/bash

# Set your PostgreSQL connection variables
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-product_api}
DB_USERNAME=${DB_USERNAME:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_SSL_MODE=${DB_SSL_MODE:-disable}

# Function to print usage
function print_usage {
    echo "Database Migration Script"
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --up                 Apply migrations (default)"
    echo "  --down               Rollback migrations"
    echo "  --migration=NAME     Apply/rollback a specific migration"
    echo "  --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --up              Apply all pending migrations"
    echo "  $0 --down            Rollback the most recent migration"
    echo "  $0 --migration=001_initial_schema       Apply/rollback a specific migration"
}

# Parse command line arguments
UP=true
DOWN=false
MIGRATION=""

for arg in "$@"
do
    case $arg in
        --up)
        UP=true
        DOWN=false
        shift
        ;;
        --down)
        UP=false
        DOWN=true
        shift
        ;;
        --migration=*)
        MIGRATION="${arg#*=}"
        shift
        ;;
        --help)
        print_usage
        exit 0
        ;;
        *)
        echo "Unknown option: $arg"
        print_usage
        exit 1
        ;;
    esac
done

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Get the project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Build the migration tool
echo "Building migration tool..."
go build -o "$SCRIPT_DIR/migrate_tool" "$PROJECT_ROOT/cmd/migrate/main.go"

# Set migration command options
CMD="$SCRIPT_DIR/migrate_tool"
if [ "$DOWN" = true ]; then
    CMD="$CMD -down"
fi

if [ ! -z "$MIGRATION" ]; then
    CMD="$CMD -migration=$MIGRATION"
fi

# Run the migration
echo "Running migrations..."
$CMD

# Remove the migration binary
rm "$SCRIPT_DIR/migrate_tool"

echo "Migration complete!"

go get github.com/gorilla/websocket