#!/bin/bash

# Generate Config Script
# This script generates the actual config by replacing placeholders with real values

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -p, --private-key KEY     Private key for blockchain account"
    echo "  -i, --identity-key KEY    Identity private key for libp2p"
    echo "  -d, --database-url URL    Database connection URL"
    echo "  -r, --redis-password PASS Redis password"
    echo "  -o, --output FILE         Output file (default: generated-config.yaml)"
    echo "  -t, --template FILE       Template file (default: config.yaml)"
    echo "  --help                    Show this help message"
    echo
    echo "Examples:"
    echo "  $0 -p 'your-private-key' -i 'your-identity-key'"
    echo "  $0 --private-key 'key' --database-url 'postgresql://user:pass@host:5432/db'"
}

# Parse command line arguments
PRIVATE_KEY=""
IDENTITY_KEY=""
DATABASE_URL=""
REDIS_PASSWORD=""
OUTPUT_FILE="generated-config.yaml"
TEMPLATE_FILE="config.yaml"

while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--private-key)
            PRIVATE_KEY="$2"
            shift 2
            ;;
        -i|--identity-key)
            IDENTITY_KEY="$2"
            shift 2
            ;;
        -d|--database-url)
            DATABASE_URL="$2"
            shift 2
            ;;
        -r|--redis-password)
            REDIS_PASSWORD="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -t|--template)
            TEMPLATE_FILE="$2"
            shift 2
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if template file exists
if [[ ! -f "$TEMPLATE_FILE" ]]; then
    print_error "Template file not found: $TEMPLATE_FILE"
    exit 1
fi

print_status "Generating config from template: $TEMPLATE_FILE"

# Read template content
TEMPLATE_CONTENT=$(cat "$TEMPLATE_FILE")

# Replace placeholders with actual values
if [[ -n "$PRIVATE_KEY" ]]; then
    TEMPLATE_CONTENT=$(echo "$TEMPLATE_CONTENT" | sed "s/private_key: \"\"/private_key: \"$PRIVATE_KEY\"/")
    print_status "Private key configured"
else
    print_warning "Private key not provided, using placeholder"
fi

if [[ -n "$IDENTITY_KEY" ]]; then
    TEMPLATE_CONTENT=$(echo "$TEMPLATE_CONTENT" | sed "s/privkey: \"\"/privkey: \"$IDENTITY_KEY\"/")
    print_status "Identity private key configured"
else
    print_warning "Identity private key not provided, using placeholder"
fi

if [[ -n "$DATABASE_URL" ]]; then
    TEMPLATE_CONTENT=$(echo "$TEMPLATE_CONTENT" | sed "s|host: postgres|host: $(echo $DATABASE_URL | sed 's|.*://\([^:]*\).*|\1|')|")
    print_status "Database URL configured"
else
    print_warning "Database URL not provided, using default"
fi

if [[ -n "$REDIS_PASSWORD" ]]; then
    TEMPLATE_CONTENT=$(echo "$TEMPLATE_CONTENT" | sed "s/db: 0/db: 0\n  password: \"$REDIS_PASSWORD\"/")
    print_status "Redis password configured"
else
    print_warning "Redis password not provided, using default"
fi

# Write generated config to output file
echo "$TEMPLATE_CONTENT" > "$OUTPUT_FILE"

print_success "Config generated successfully: $OUTPUT_FILE"

# Show summary
echo
print_status "Configuration Summary:"
echo "  Template: $TEMPLATE_FILE"
echo "  Output: $OUTPUT_FILE"
echo "  Private Key: ${PRIVATE_KEY:+✓ Set}${PRIVATE_KEY:-✗ Not set}"
echo "  Identity Key: ${IDENTITY_KEY:+✓ Set}${IDENTITY_KEY:-✗ Not set}"
echo "  Database URL: ${DATABASE_URL:+✓ Set}${DATABASE_URL:-✗ Not set}"
echo "  Redis Password: ${REDIS_PASSWORD:+✓ Set}${REDIS_PASSWORD:-✗ Not set}"
echo

# Instructions for next steps
print_status "Next steps:"
echo "  1. Review the generated config: cat $OUTPUT_FILE"
echo "  2. Update the ConfigMap with the generated config"
echo "  3. Apply the updated ConfigMap to Kubernetes"
echo "  4. Restart the deployment to pick up the new config" 