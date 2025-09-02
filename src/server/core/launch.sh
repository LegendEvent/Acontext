#!/bin/bash

# Acontext Core Launch Script
# Supports both development (uvicorn) and production (gunicorn) modes

set -e

# Default configuration
MODE="${MODE:-dev}"
HOST="${HOST:-0.0.0.0}"
PORT="${PORT:-8000}"
WORKERS="${WORKERS:-4}"
RELOAD="${RELOAD:-true}"
LOG_LEVEL="${LOG_LEVEL:-info}"
APP_MODULE="${APP_MODULE:-core_asgi:app}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo "${GREEN}[SUCCESS]${NC} $1"
}

show_help() {
    cat << EOF
Acontext Core Launch Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -m, --mode MODE         Run mode: 'dev' (uvicorn) or 'prod' (gunicorn) [default: dev]
    -h, --host HOST         Host to bind to [default: 0.0.0.0]
    -p, --port PORT         Port to bind to [default: 8000]
    -w, --workers WORKERS   Number of worker processes (prod mode only) [default: 4]
    -r, --reload RELOAD     Enable auto-reload (dev mode only) [default: true]
    -l, --log-level LEVEL   Log level: debug, info, warning, error [default: info]
    -a, --app APP_MODULE    ASGI application module [default: core_asgi:app]
    --help                  Show this help message

ENVIRONMENT VARIABLES:
    MODE, HOST, PORT, WORKERS, RELOAD, LOG_LEVEL, APP_MODULE
    (Command line arguments override environment variables)

EXAMPLES:
    # Development mode with auto-reload
    $0 --mode dev --port 8000 --reload true

    # Production mode with 8 workers
    $0 --mode prod --workers 8 --port 8000

    # Using environment variables
    MODE=prod WORKERS=6 PORT=8080 $0
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mode)
            MODE="$2"
            shift 2
            ;;
        -h|--host)
            HOST="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -w|--workers)
            WORKERS="$2"
            shift 2
            ;;
        -r|--reload)
            RELOAD="$2"
            shift 2
            ;;
        -l|--log-level)
            LOG_LEVEL="$2"
            shift 2
            ;;
        -a|--app)
            APP_MODULE="$2"
            shift 2
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate mode
if [[ "$MODE" != "dev" && "$MODE" != "prod" ]]; then
    log_error "Invalid mode: $MODE. Must be 'dev' or 'prod'"
    exit 1
fi

# Validate port
if ! [[ "$PORT" =~ ^[0-9]+$ ]] || [ "$PORT" -lt 1 ] || [ "$PORT" -gt 65535 ]; then
    log_error "Invalid port: $PORT. Must be between 1 and 65535"
    exit 1
fi

# Validate workers (for prod mode)
if [[ "$MODE" == "prod" ]]; then
    if ! [[ "$WORKERS" =~ ^[0-9]+$ ]] || [ "$WORKERS" -lt 1 ]; then
        log_error "Invalid workers count: $WORKERS. Must be a positive integer"
        exit 1
    fi
fi

# Validate log level
valid_log_levels=("debug" "info" "warning" "error")
if [[ ! " ${valid_log_levels[@]} " =~ " ${LOG_LEVEL} " ]]; then
    log_error "Invalid log level: $LOG_LEVEL. Must be one of: ${valid_log_levels[*]}"
    exit 1
fi

# Check if app module exists
if [[ ! -f "${APP_MODULE%%:*}.py" ]]; then
    log_warn "App module file ${APP_MODULE%%:*}.py not found in current directory"
fi

# Display configuration
log_info "Starting Acontext Core..."
log_info "Mode: $MODE"
log_info "Host: $HOST"
log_info "Port: $PORT"
log_info "App Module: $APP_MODULE"
log_info "Log Level: $LOG_LEVEL"

if [[ "$MODE" == "dev" ]]; then
    log_info "Auto-reload: $RELOAD"
elif [[ "$MODE" == "prod" ]]; then
    log_info "Workers: $WORKERS"
fi

echo ""

# Launch based on mode
if [[ "$MODE" == "dev" ]]; then
    log_success "Starting development server with uvicorn..."
    
    # Build uvicorn command
    UVICORN_CMD="uv run -m uvicorn $APP_MODULE --host $HOST --port $PORT --log-level $LOG_LEVEL"
    
    if [[ "$RELOAD" == "true" ]]; then
        UVICORN_CMD="$UVICORN_CMD --reload"
    fi
    
    log_info "Command: $UVICORN_CMD"
    exec $UVICORN_CMD
    
elif [[ "$MODE" == "prod" ]]; then
    log_success "Starting production server with gunicorn..."
    
    # Build gunicorn command
    GUNICORN_CMD="uv run -m gunicorn $APP_MODULE"
    GUNICORN_CMD="$GUNICORN_CMD --bind $HOST:$PORT"
    GUNICORN_CMD="$GUNICORN_CMD --workers $WORKERS"
    GUNICORN_CMD="$GUNICORN_CMD --worker-class uvicorn.workers.UvicornWorker"
    GUNICORN_CMD="$GUNICORN_CMD --log-level $LOG_LEVEL"
    GUNICORN_CMD="$GUNICORN_CMD --access-logfile -"
    GUNICORN_CMD="$GUNICORN_CMD --error-logfile -"
    
    log_info "Command: $GUNICORN_CMD"
    exec $GUNICORN_CMD
fi
