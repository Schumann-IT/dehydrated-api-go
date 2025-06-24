#!/bin/bash

# Azure AD Setup Script for Dehydrated API Go using Custom API
# This script sets up Azure AD authentication using a custom API application
# with defined app roles and proper permissions

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
    
    # Check for existing API application
    EXISTING_API_APP=$(az ad app list --display-name "Dehydrated API Go" --query "[0].appId" -o tsv)
    EXISTING_CLIENT_APP=$(az ad app list --display-name "Dehydrated API Client" --query "[0].appId" -o tsv)
    
    if [ "$EXISTING_API_APP" != "null" ] && [ -n "$EXISTING_API_APP" ]; then
        print_warning "Found existing API application: $EXISTING_API_APP"
        read -p "Do you want to remove the existing API application? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "Removing existing API application..."
            
            # Remove service principal first
            EXISTING_API_SP=$(az ad sp list --display-name "Dehydrated API Go" --query "[0].id" -o tsv)
            if [ "$EXISTING_API_SP" != "null" ] && [ -n "$EXISTING_API_SP" ]; then
                az ad sp delete --id $EXISTING_API_SP
                print_success "Removed existing API service principal"
            fi
            
            # Remove application
            az ad app delete --id $EXISTING_API_APP
            print_success "Removed existing API application"
        else
            print_status "Using existing API application: $EXISTING_API_APP"
            API_APP_ID=$EXISTING_API_APP
        fi
    fi
    
    if [ "$EXISTING_CLIENT_APP" != "null" ] && [ -n "$EXISTING_CLIENT_APP" ]; then
        print_warning "Found existing client application: $EXISTING_CLIENT_APP"
        read -p "Do you want to remove the existing client application? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "Removing existing client application..."
            
            # Remove service principal first
            EXISTING_CLIENT_SP=$(az ad sp list --display-name "Dehydrated API Client" --query "[0].id" -o tsv)
            if [ "$EXISTING_CLIENT_SP" != "null" ] && [ -n "$EXISTING_CLIENT_SP" ]; then
                az ad sp delete --id $EXISTING_CLIENT_SP
                print_success "Removed existing client service principal"
            fi
            
            # Remove application
            az ad app delete --id $EXISTING_CLIENT_APP
            print_success "Removed existing client application"
        else
            print_status "Using existing client application: $EXISTING_CLIENT_APP"
            CLIENT_APP_ID=$EXISTING_CLIENT_APP
        fi
    fi
}

# Function to create API application
create_api_app() {
    print_status "Creating API application..."
    
    # Check if API app already exists (from cleanup or previous run)
    if [ -n "$API_APP_ID" ]; then
        print_warning "Using existing API application: $API_APP_ID"
        return 0
    fi
    
    EXISTING_APP=$(az ad app list --display-name "Dehydrated API Go" --query "[0].appId" -o tsv)
    
    if [ "$EXISTING_APP" != "null" ] && [ -n "$EXISTING_APP" ]; then
        print_warning "API application already exists with ID: $EXISTING_APP"
        API_APP_ID=$EXISTING_APP
    else
        # Create new API application
        API_APP_ID=$(az ad app create \
            --display-name "Dehydrated API Go" \
            --sign-in-audience "AzureADMyOrg" \
            --enable-access-token-issuance true \
            --enable-id-token-issuance false \
            --query "appId" -o tsv)
        
        print_success "Created API application with ID: $API_APP_ID"
    fi
    
    # Add identifier URI
    print_status "Adding identifier URI to API application..."
    az ad app update \
        --id $API_APP_ID \
        --identifier-uris "https://$TENANT_NAME/$API_APP_ID"
    
    print_success "Added identifier URI: https://$TENANT_NAME/$API_APP_ID"
}

# Function to create app roles for the API
create_app_roles() {
    print_status "Creating app roles for API application..."
    
    # Create app roles JSON
    cat > app-roles.json << EOF
[
  {
    "allowedMemberTypes": ["User", "Application"],
    "displayName": "Access Dehydrated API",
    "description": "Allows access to the Dehydrated API Go service",
    "value": "user_impersonation",
    "isEnabled": true
  },
  {
    "allowedMemberTypes": ["Application"],
    "displayName": "Service Access",
    "description": "Allows service-to-service access to the Dehydrated API Go service",
    "value": "service_access",
    "isEnabled": true
  }
]
EOF
    
    # Add app roles to the API application
    az ad app update \
        --id $API_APP_ID \
        --app-roles @app-roles.json
    
    # Clean up temporary file
    rm app-roles.json
    
    print_success "Created app roles for API application"
}

# Function to create client application
create_client_app() {
    print_status "Creating client application..."
    
    # Check if client app already exists (from cleanup or previous run)
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

# Function to create service principals
create_service_principals() {
    print_status "Creating service principals..."
    
    # Create service principal for API
    EXISTING_API_SP=$(az ad sp list --display-name "Dehydrated API Go" --query "[0].id" -o tsv)
    if [ "$EXISTING_API_SP" != "null" ] && [ -n "$EXISTING_API_SP" ]; then
        print_warning "API service principal already exists"
    else
        az ad sp create --id $API_APP_ID
        print_success "Created API service principal"
    fi
    
    # Create service principal for client
    EXISTING_CLIENT_SP=$(az ad sp list --display-name "Dehydrated API Client" --query "[0].id" -o tsv)
    if [ "$EXISTING_CLIENT_SP" != "null" ] && [ -n "$EXISTING_CLIENT_SP" ]; then
        print_warning "Client service principal already exists"
    else
        az ad sp create --id $CLIENT_APP_ID
        print_success "Created client service principal"
    fi
}

# Function to grant API permissions
grant_api_permissions() {
    print_status "Granting API permissions..."
    
    # Get the app role IDs from the API application
    USER_IMPERSONATION_ROLE_ID=$(az ad app show --id $API_APP_ID --query "appRoles[?value=='user_impersonation'].id" -o tsv)
    SERVICE_ACCESS_ROLE_ID=$(az ad app show --id $API_APP_ID --query "appRoles[?value=='service_access'].id" -o tsv)
    
    if [ -z "$USER_IMPERSONATION_ROLE_ID" ]; then
        print_error "Failed to get user_impersonation role ID from API application"
        exit 1
    fi
    
    print_status "User impersonation role ID: $USER_IMPERSONATION_ROLE_ID"
    if [ -n "$SERVICE_ACCESS_ROLE_ID" ]; then
        print_status "Service access role ID: $SERVICE_ACCESS_ROLE_ID"
    fi
    
    # Check if permission already exists
    EXISTING_PERMISSION=$(az ad app permission list --id $CLIENT_APP_ID --query "[?resourceAppId=='$API_APP_ID'].resourceAppId" -o tsv)
    
    if [ "$EXISTING_PERMISSION" = "$API_APP_ID" ]; then
        print_warning "API permission already exists"
    else
        # Grant API permissions using the role ID
        az ad app permission add \
            --id $CLIENT_APP_ID \
            --api $API_APP_ID \
            --api-permissions "$USER_IMPERSONATION_ROLE_ID=Role"
        
        print_success "Granted API permissions"
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
    
    # Note: Azure CLI cannot directly access custom APIs without proper configuration
    # We'll test if the API application is properly configured instead
    
    print_status "Checking API application configuration..."
    
    # Verify the API application exists and has the correct configuration
    API_APP_CHECK=$(az ad app show --id $API_APP_ID --query "appId" -o tsv 2>/dev/null)
    
    if [ "$API_APP_CHECK" = "$API_APP_ID" ]; then
        print_success "API application is properly configured"
        
        # Check if app roles are configured
        APP_ROLES_COUNT=$(az ad app show --id $API_APP_ID --query "appRoles | length(@)" -o tsv)
        if [ "$APP_ROLES_COUNT" -gt 0 ]; then
            print_success "API application has $APP_ROLES_COUNT app roles configured"
        else
            print_warning "API application has no app roles configured"
        fi
        
        # Check if client has permissions
        CLIENT_PERMISSIONS=$(az ad app permission list --id $CLIENT_APP_ID --query "[?resourceAppId=='$API_APP_ID'].resourceAppId" -o tsv)
        if [ "$CLIENT_PERMISSIONS" = "$API_APP_ID" ]; then
            print_success "Client application has permissions to API"
        else
            print_warning "Client application may not have proper permissions to API"
        fi
        
    else
        print_error "API application is not properly configured"
        exit 1
    fi
    
    print_warning "Note: Azure CLI cannot directly test custom API tokens without additional configuration."
    print_status "You can test token generation after starting your application using:"
    echo "  1. Start your application:"
    echo "     go run cmd/api/main.go --config examples/config-auth.yaml"
    echo ""
    echo "  2. Use a tool like Postman or curl with a token from your application's OAuth flow"
    echo ""
    echo "  3. Or use the test script with a manually obtained token:"
    echo "     ./scripts/test-auth.sh --api-url http://localhost:3000"
    echo ""
}

# Function to generate configuration file
generate_config() {
    print_status "Generating configuration file..."
    
    CONFIG_FILE="config-auth-custom.yaml"
    
    cat > $CONFIG_FILE << EOF
port: 3000
dehydratedBaseDir: ./data
enableWatcher: false
logging:
  level: debug
  encoding: console
  outputPath: ""

# Azure AD authentication configuration using Custom API
auth:
  # The Azure AD tenant ID
  tenantId: "$TENANT_ID"
  # The API application ID (resource)
  clientId: "$API_APP_ID"
  # The authority URL for Azure AD
  authority: "https://login.microsoftonline.com/$TENANT_ID"
  # List of allowed audience values in the JWT token
  allowedAudiences:
    - "https://$TENANT_NAME/$API_APP_ID"
    - "api://$API_APP_ID"

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
    echo "Azure AD Custom API Setup Complete!"
    echo "=========================================="
    echo ""
    echo "Configuration:"
    echo "  Tenant ID: $TENANT_ID"
    echo "  Tenant Name: $TENANT_NAME"
    echo "  API App ID: $API_APP_ID"
    echo "  Client App ID: $CLIENT_APP_ID"
    echo "  Resource: Custom API"
    echo ""
    echo "Next Steps:"
    echo "1. Copy the generated config file:"
    echo "   cp $CONFIG_FILE examples/config-auth.yaml"
    echo ""
    echo "2. Start your application:"
    echo "   go run cmd/api/main.go --config examples/config-auth.yaml"
    echo ""
    echo "3. Create a client secret for token generation:"
    echo "   ./scripts/get-custom-api-token.sh --create-secret --client-id $CLIENT_APP_ID"
    echo ""
    echo "4. Generate and test access tokens:"
    echo "   ./scripts/get-custom-api-token.sh \\"
    echo "     --tenant-id $TENANT_ID \\"
    echo "     --client-id $CLIENT_APP_ID \\"
    echo "     --client-secret YOUR_SECRET \\"
    echo "     --resource https://$TENANT_NAME/$API_APP_ID \\"
    echo "     --test-api"
    echo ""
    echo "5. For development, you can also use:"
    echo "   ./scripts/get-custom-api-token.sh \\"
    echo "     --tenant-id $TENANT_ID \\"
    echo "     --client-id $CLIENT_APP_ID \\"
    echo "     --client-secret YOUR_SECRET \\"
    echo "     --resource https://$TENANT_NAME/$API_APP_ID \\"
    echo "     --output token"
    echo ""
    echo "App Roles Available:"
    echo "  - user_impersonation: For user authentication"
    echo "  - service_access: For service-to-service authentication"
    echo ""
    echo "Note: Azure CLI cannot directly access custom APIs. Use the token generator script instead."
    echo ""
}

# Function to display cleanup instructions
display_cleanup_instructions() {
    echo ""
    echo "=========================================="
    echo "Cleanup Instructions"
    echo "=========================================="
    echo ""
    echo "To remove the created applications:"
    echo ""
    echo "1. Remove service principals:"
    echo "   az ad sp delete --id \$(az ad sp list --display-name 'Dehydrated API Go' --query '[0].id' -o tsv)"
    echo "   az ad sp delete --id \$(az ad sp list --display-name 'Dehydrated API Client' --query '[0].id' -o tsv)"
    echo ""
    echo "2. Remove applications:"
    echo "   az ad app delete --id $API_APP_ID"
    echo "   az ad app delete --id $CLIENT_APP_ID"
    echo ""
    echo "Or simply run this script again and choose to remove existing applications."
    echo ""
}

# Main execution
main() {
    echo "=========================================="
    echo "Azure AD Setup for Dehydrated API Go"
    echo "Using Custom API"
    echo "=========================================="
    echo ""
    
    check_azure_cli
    get_tenant_info
    cleanup_existing_apps
    create_api_app
    create_app_roles
    create_client_app
    create_service_principals
    grant_api_permissions
    test_token_generation
    generate_config
    display_instructions
    display_cleanup_instructions
}

# Run main function
main "$@" 