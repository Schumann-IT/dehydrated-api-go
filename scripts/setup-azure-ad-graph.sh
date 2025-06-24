#!/bin/bash

# Azure AD Setup Script for Dehydrated API Go using Microsoft Graph
# This script sets up Azure AD authentication using Microsoft Graph API
# which is simpler than creating custom API applications

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to check if Azure CLI is installed and logged in
check_azure_cli() {
    print_status "Checking Azure CLI installation..."
    
    if ! command -v az &> /dev/null; then
        print_error "Azure CLI is not installed. Please install it first:"
        echo "  https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
        exit 1
    fi
    
    print_success "Azure CLI is installed"
    
    print_status "Checking Azure login status..."
    if ! az account show &> /dev/null; then
        print_error "Not logged in to Azure. Please run 'az login' first"
        exit 1
    fi
    
    print_success "Logged in to Azure"
}

# Function to get tenant information
get_tenant_info() {
    print_status "Getting tenant information..."
    
    TENANT_ID=$(az account show --query "tenantId" -o tsv)
    TENANT_NAME=$(az account show --query "user.name" -o tsv | cut -d'@' -f2)
    SUBSCRIPTION_ID=$(az account show --query "id" -o tsv)
    
    if [ -z "$TENANT_ID" ] || [ -z "$TENANT_NAME" ]; then
        print_error "Failed to get tenant information"
        exit 1
    fi
    
    print_success "Tenant ID: $TENANT_ID"
    print_success "Tenant Name: $TENANT_NAME"
    print_success "Subscription ID: $SUBSCRIPTION_ID"
}

# Function to cleanup existing applications
cleanup_existing_apps() {
    print_status "Checking for existing applications..."
    
    # Check for existing client application
    EXISTING_CLIENT_APP=$(az ad app list --display-name "Dehydrated API Client" --query "[0].appId" -o tsv)
    
    if [ "$EXISTING_CLIENT_APP" != "null" ] && [ -n "$EXISTING_CLIENT_APP" ]; then
        print_warning "Found existing client application: $EXISTING_CLIENT_APP"
        read -p "Do you want to remove the existing application? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "Removing existing client application..."
            
            # Remove service principal first
            EXISTING_SP=$(az ad sp list --display-name "Dehydrated API Client" --query "[0].id" -o tsv)
            if [ "$EXISTING_SP" != "null" ] && [ -n "$EXISTING_SP" ]; then
                az ad sp delete --id $EXISTING_SP
                print_success "Removed existing service principal"
            fi
            
            # Remove application
            az ad app delete --id $EXISTING_CLIENT_APP
            print_success "Removed existing client application"
        else
            print_status "Using existing client application: $EXISTING_CLIENT_APP"
            CLIENT_APP_ID=$EXISTING_CLIENT_APP
            return 0
        fi
    fi
}

# Function to create client application
create_client_app() {
    print_status "Creating client application..."
    
    # Check if app already exists (from cleanup or previous run)
    if [ -n "$CLIENT_APP_ID" ]; then
        print_warning "Using existing client application: $CLIENT_APP_ID"
        return 0
    fi
    
    EXISTING_APP=$(az ad app list --display-name "Dehydrated API Client" --query "[0].appId" -o tsv)
    
    if [ "$EXISTING_APP" != "null" ] && [ -n "$EXISTING_APP" ]; then
        print_warning "Client application already exists with ID: $EXISTING_APP"
        CLIENT_APP_ID=$EXISTING_APP
    else
        # Create new client application
        CLIENT_APP_ID=$(az ad app create \
            --display-name "Dehydrated API Client" \
            --sign-in-audience "AzureADMyOrg" \
            --enable-access-token-issuance false \
            --enable-id-token-issuance false \
            --query "appId" -o tsv)
        
        print_success "Created client application with ID: $CLIENT_APP_ID"
    fi
}

# Function to create service principal
create_service_principal() {
    print_status "Creating service principal for client application..."
    
    # Check if service principal already exists
    EXISTING_SP=$(az ad sp list --display-name "Dehydrated API Client" --query "[0].id" -o tsv)
    
    if [ "$EXISTING_SP" != "null" ] && [ -n "$EXISTING_SP" ]; then
        print_warning "Service principal already exists"
    else
        az ad sp create --id $CLIENT_APP_ID
        print_success "Created service principal"
    fi
}

# Function to grant Microsoft Graph permissions
grant_graph_permissions() {
    print_status "Granting Microsoft Graph permissions..."
    
    # Microsoft Graph resource ID
    GRAPH_RESOURCE_ID="00000003-0000-0000-c000-000000000000"
    
    # Check if permission already exists
    EXISTING_PERMISSION=$(az ad app permission list --id $CLIENT_APP_ID --query "[?resourceAppId=='$GRAPH_RESOURCE_ID'].resourceAppId" -o tsv)
    
    if [ "$EXISTING_PERMISSION" = "$GRAPH_RESOURCE_ID" ]; then
        print_warning "Microsoft Graph permission already exists"
    else
        # Get the User.Read permission ID from Microsoft Graph
        USER_READ_PERMISSION_ID=$(az ad sp show --id $GRAPH_RESOURCE_ID --query "oauth2PermissionScopes[?value=='User.Read'].id" -o tsv)
        
        if [ -z "$USER_READ_PERMISSION_ID" ]; then
            print_error "Failed to get User.Read permission ID from Microsoft Graph"
            exit 1
        fi
        
        print_status "User.Read permission ID: $USER_READ_PERMISSION_ID"
        
        # Grant Microsoft Graph permissions using the permission ID
        az ad app permission add \
            --id $CLIENT_APP_ID \
            --api $GRAPH_RESOURCE_ID \
            --api-permissions "$USER_READ_PERMISSION_ID=Scope"
        
        print_success "Granted Microsoft Graph permissions"
    fi
    
    # Grant admin consent with retry logic
    print_status "Granting admin consent..."
    MAX_RETRIES=3
    RETRY_COUNT=0
    
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if az ad app permission admin-consent --id $CLIENT_APP_ID 2>/dev/null; then
            print_success "Admin consent granted"
            break
        else
            RETRY_COUNT=$((RETRY_COUNT + 1))
            if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
                print_warning "Admin consent failed, retrying in 5 seconds... (attempt $RETRY_COUNT/$MAX_RETRIES)"
                sleep 5
            else
                print_warning "Admin consent failed after $MAX_RETRIES attempts. You may need to grant consent manually."
                print_status "You can grant consent manually by visiting:"
                echo "  https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps/ApplicationMenuBlade/CallAnAPI/appId/$CLIENT_APP_ID"
            fi
        fi
    done
}

# Function to test token generation
test_token_generation() {
    print_status "Testing token generation..."
    
    TOKEN=$(az account get-access-token --resource "https://graph.microsoft.com" --query "accessToken" -o tsv)
    
    if [ -n "$TOKEN" ]; then
        print_success "Successfully generated access token"
        
        # Decode and display token information (without sensitive data)
        print_status "Token information:"
        echo $TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null | jq '. | {aud, iss, exp, iat}' 2>/dev/null || echo "Token decoded successfully"
    else
        print_error "Failed to generate access token"
        exit 1
    fi
}

# Function to generate configuration file
generate_config() {
    print_status "Generating configuration file..."
    
    CONFIG_FILE="config-auth-graph.yaml"
    
    cat > $CONFIG_FILE << EOF
port: 3000
dehydratedBaseDir: ./data
enableWatcher: false
logging:
  level: debug
  encoding: console
  outputPath: ""

# Azure AD authentication configuration using Microsoft Graph
auth:
  # The Azure AD tenant ID
  tenantId: "$TENANT_ID"
  # Microsoft Graph resource ID
  clientId: "00000003-0000-0000-c000-000000000000"
  # The authority URL for Azure AD
  authority: "https://login.microsoftonline.com/$TENANT_ID"
  # List of allowed audience values in the JWT token
  allowedAudiences:
    - "https://graph.microsoft.com"
    - "https://$TENANT_NAME/00000003-0000-0000-c000-000000000000"

plugins:
  example:
    enabled: true
    registry:
      type: local
      config:
        path: ./examples/plugins/simple/simple
    config:
      name: example
EOF
    
    print_success "Configuration file generated: $CONFIG_FILE"
}

# Function to display usage instructions
display_instructions() {
    echo ""
    echo "=========================================="
    echo "Azure AD Setup Complete!"
    echo "=========================================="
    echo ""
    echo "Configuration:"
    echo "  Tenant ID: $TENANT_ID"
    echo "  Tenant Name: $TENANT_NAME"
    echo "  Client App ID: $CLIENT_APP_ID"
    echo "  Resource: Microsoft Graph"
    echo ""
    echo "Next Steps:"
    echo "1. Copy the generated config file:"
    echo "   cp $CONFIG_FILE examples/config-auth.yaml"
    echo ""
    echo "2. Start your application:"
    echo "   go run cmd/api/main.go --config examples/config-auth.yaml"
    echo ""
    echo "3. Test authentication:"
    echo "   TOKEN=\$(az account get-access-token --resource 'https://graph.microsoft.com' --query 'accessToken' -o tsv)"
    echo "   curl -H \"Authorization: Bearer \$TOKEN\" http://localhost:3000/api/v1/domains"
    echo ""
    echo "4. For development, you can also use:"
    echo "   az account get-access-token --resource 'https://graph.microsoft.com'"
    echo ""
}

# Main execution
main() {
    echo "=========================================="
    echo "Azure AD Setup for Dehydrated API Go"
    echo "Using Microsoft Graph API"
    echo "=========================================="
    echo ""
    
    check_azure_cli
    get_tenant_info
    cleanup_existing_apps
    create_client_app
    create_service_principal
    grant_graph_permissions
    test_token_generation
    generate_config
    display_instructions
}

# Run main function
main "$@" 