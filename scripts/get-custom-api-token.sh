#!/bin/bash

# Script to generate access tokens for custom Azure AD APIs
# This script uses the OAuth2 client credentials flow to get tokens for custom APIs

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

# Function to display usage
display_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --tenant-id ID       Azure AD tenant ID"
    echo "  --client-id ID       Client application ID"
    echo "  --client-secret SECRET Client application secret"
    echo "  --resource URI       Resource URI (e.g., https://tenant.com/api-id)"
    echo "  --scope SCOPE        OAuth scope (default: .default)"
    echo "  --output FORMAT      Output format: token, json, or full (default: token)"
    echo "  --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --tenant-id YOUR_TENANT_ID --client-id YOUR_CLIENT_ID --client-secret YOUR_SECRET --resource https://tenant.com/api-id"
    echo "  $0 --tenant-id YOUR_TENANT_ID --client-id YOUR_CLIENT_ID --client-secret YOUR_SECRET --resource https://tenant.com/api-id --output json"
    echo ""
    echo "Note: You need to create a client secret for your client application first."
    echo "You can create one using:"
    echo "  az ad app credential reset --id YOUR_CLIENT_ID --append"
}

# Function to check dependencies
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed"
        exit 1
    fi
}

# Function to create client secret
create_client_secret() {
    local client_id=$1
    
    print_status "Creating client secret for application: $client_id"
    
    # Create a new client secret
    SECRET_RESPONSE=$(az ad app credential reset --id $client_id --append --query "{password: password, endDate: endDate}" -o json)
    
    if [ $? -eq 0 ]; then
        CLIENT_SECRET=$(echo $SECRET_RESPONSE | jq -r '.password')
        END_DATE=$(echo $SECRET_RESPONSE | jq -r '.endDate')
        
        print_success "Created client secret (expires: $END_DATE)"
        print_warning "Store this secret securely! It won't be shown again."
        echo "Client Secret: $CLIENT_SECRET"
        echo ""
        
        return 0
    else
        print_error "Failed to create client secret"
        return 1
    fi
}

# Function to get access token
get_access_token() {
    local tenant_id=$1
    local client_id=$2
    local client_secret=$3
    local resource=$4
    local scope=${5:-".default"}
    
    print_status "Getting access token for resource: $resource"
    
    # Construct the token endpoint URL
    TOKEN_URL="https://login.microsoftonline.com/$tenant_id/oauth2/v2.0/token"
    
    # Prepare the request body
    REQUEST_BODY="grant_type=client_credentials&client_id=$client_id&client_secret=$client_secret&scope=$resource/$scope"
    
    # Make the request
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "$REQUEST_BODY" \
        "$TOKEN_URL")
    
    # Check if the request was successful
    if echo "$RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
        print_success "Successfully obtained access token"
        
        # Extract token based on output format
        case "$OUTPUT_FORMAT" in
            "token")
                echo "$RESPONSE" | jq -r '.access_token'
                ;;
            "json")
                echo "$RESPONSE" | jq '.'
                ;;
            "full")
                echo "$RESPONSE" | jq '.'
                echo ""
                print_status "Token information:"
                TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
                echo $TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null | jq '. | {aud, iss, exp, iat, appid}' 2>/dev/null || echo "Token decoded successfully"
                ;;
            *)
                echo "$RESPONSE" | jq -r '.access_token'
                ;;
        esac
        
        return 0
    else
        print_error "Failed to get access token"
        echo "Response: $RESPONSE"
        return 1
    fi
}

# Function to test the token with your API
test_token_with_api() {
    local token=$1
    local api_url=${2:-"http://localhost:3000"}
    local endpoint=${3:-"/api/v1/domains"}
    
    print_status "Testing token with API: $api_url$endpoint"
    
    # Check if API is running
    if ! curl -s --connect-timeout 5 "$api_url/health" &> /dev/null; then
        print_warning "API server doesn't seem to be running at $api_url"
        print_status "Please start the server first:"
        echo "  go run cmd/api/main.go --config examples/config-auth.yaml"
        return 1
    fi
    
    # Test with token
    RESPONSE=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        "$api_url$endpoint")
    
    HTTP_CODE="${RESPONSE: -3}"
    RESPONSE_BODY="${RESPONSE%???}"
    
    if [ "$HTTP_CODE" = "200" ]; then
        print_success "API request successful (HTTP $HTTP_CODE)"
        echo "Response: $RESPONSE_BODY"
        return 0
    else
        print_error "API request failed (HTTP $HTTP_CODE)"
        echo "Response: $RESPONSE_BODY"
        return 1
    fi
}

# Parse command line arguments
TENANT_ID=""
CLIENT_ID=""
CLIENT_SECRET=""
RESOURCE=""
SCOPE=".default"
OUTPUT_FORMAT="token"
CREATE_SECRET=false
TEST_API=false
API_URL="http://localhost:3000"
API_ENDPOINT="/api/v1/domains"

while [[ $# -gt 0 ]]; do
    case $1 in
        --tenant-id)
            TENANT_ID="$2"
            shift 2
            ;;
        --client-id)
            CLIENT_ID="$2"
            shift 2
            ;;
        --client-secret)
            CLIENT_SECRET="$2"
            shift 2
            ;;
        --resource)
            RESOURCE="$2"
            shift 2
            ;;
        --scope)
            SCOPE="$2"
            shift 2
            ;;
        --output)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        --create-secret)
            CREATE_SECRET=true
            shift
            ;;
        --test-api)
            TEST_API=true
            shift
            ;;
        --api-url)
            API_URL="$2"
            shift 2
            ;;
        --api-endpoint)
            API_ENDPOINT="$2"
            shift 2
            ;;
        --help)
            display_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            display_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    echo "=========================================="
    echo "Custom API Token Generator"
    echo "=========================================="
    echo ""
    
    check_dependencies
    
    # Validate required parameters
    if [ -z "$TENANT_ID" ] || [ -z "$CLIENT_ID" ] || [ -z "$RESOURCE" ]; then
        print_error "Missing required parameters"
        display_usage
        exit 1
    fi
    
    # Create client secret if requested
    if [ "$CREATE_SECRET" = true ]; then
        if create_client_secret "$CLIENT_ID"; then
            print_status "Please use the generated secret with --client-secret"
            exit 0
        else
            exit 1
        fi
    fi
    
    # Check if client secret is provided
    if [ -z "$CLIENT_SECRET" ]; then
        print_error "Client secret is required. Use --client-secret or --create-secret"
        exit 1
    fi
    
    # Get access token
    if get_access_token "$TENANT_ID" "$CLIENT_ID" "$CLIENT_SECRET" "$RESOURCE" "$SCOPE"; then
        TOKEN=$(get_access_token "$TENANT_ID" "$CLIENT_ID" "$CLIENT_SECRET" "$RESOURCE" "$SCOPE")
        
        # Test with API if requested
        if [ "$TEST_API" = true ]; then
            echo ""
            test_token_with_api "$TOKEN" "$API_URL" "$API_ENDPOINT"
        fi
    else
        exit 1
    fi
}

# Run main function
main "$@" 