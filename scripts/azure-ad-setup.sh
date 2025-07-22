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

# Global variables
DELETE_ONLY=false
FORCE_RECREATE=false

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

# Function to display usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --delete   Only delete existing applications (no creation)"
    echo "  --force    Delete existing applications and recreate them (no prompts)"
    echo "  -h, --help Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Check for existing apps and prompt for deletion"
    echo "  $0 --delete          # Only delete existing applications (with prompts)"
    echo "  $0 --force           # Delete existing apps and recreate them (no prompts)"
    echo "  $0 --delete --force  # Delete existing applications without prompts (no creation)"
    echo ""
}

# Function to parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --delete)
                DELETE_ONLY=true
                shift
                ;;
            --force)
                FORCE_RECREATE=true
                shift
                ;;
            -h|--help)
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
    
    # Check for conflicting options - allow --delete --force combination
    # --delete --force means delete without prompts (no creation)
    # --force alone means delete and recreate without prompts
    # --delete alone means delete with prompts (no creation)
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
    
    # Track if we found any existing applications
    FOUND_EXISTING=false
    
    if [ "$EXISTING_API_APP" != "null" ] && [ -n "$EXISTING_API_APP" ]; then
        FOUND_EXISTING=true
        print_warning "Found existing API application: $EXISTING_API_APP"
        
        if [ "$FORCE_RECREATE" = true ] || ([ "$DELETE_ONLY" = true ] && [ "$FORCE_RECREATE" = true ]); then
            print_status "Automatically removing existing API application..."
        elif [ "$DELETE_ONLY" = true ]; then
            read -p "Do you want to remove the existing API application? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_status "Using existing API application: $EXISTING_API_APP"
                API_APP_ID=$EXISTING_API_APP
                return
            fi
        else
            read -p "Do you want to remove the existing API application? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_status "Using existing API application: $EXISTING_API_APP"
                API_APP_ID=$EXISTING_API_APP
                return
            fi
        fi
        
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
    fi
    
    if [ "$EXISTING_CLIENT_APP" != "null" ] && [ -n "$EXISTING_CLIENT_APP" ]; then
        FOUND_EXISTING=true
        print_warning "Found existing client application: $EXISTING_CLIENT_APP"
        
        if [ "$FORCE_RECREATE" = true ] || ([ "$DELETE_ONLY" = true ] && [ "$FORCE_RECREATE" = true ]); then
            print_status "Automatically removing existing client application..."
        elif [ "$DELETE_ONLY" = true ]; then
            read -p "Do you want to remove the existing client application? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_status "Using existing client application: $EXISTING_CLIENT_APP"
                CLIENT_APP_ID=$EXISTING_CLIENT_APP
                return
            fi
        else
            read -p "Do you want to remove the existing client application? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_status "Using existing client application: $EXISTING_CLIENT_APP"
                CLIENT_APP_ID=$EXISTING_CLIENT_APP
                return
            fi
        fi
        
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
    fi
    
    # If no existing applications found
    if [ "$FOUND_EXISTING" = false ]; then
        print_success "No existing applications found"
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
    
    # Add identifier URIs (both formats for compatibility)
    print_status "Adding identifier URIs to API application..."
    az ad app update \
        --id $API_APP_ID \
        --identifier-uris "api://$API_APP_ID"
    
    print_success "Added identifier URIs:"
    print_success "  - api://$API_APP_ID"
}

# Function to create delegated permission for the API
create_delegated_permission() {
    print_status "Creating delegated permission for API application..."
    
    # Create delegated permission JSON
    cat > delegated-permission.json << EOF
{
  "api": {
    "oauth2PermissionScopes": [
      {
        "adminConsentDescription": "Allow the application to access the Dehydrated API on behalf of the signed-in user",
        "adminConsentDisplayName": "Access Dehydrated API",
        "id": "7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d",
        "isEnabled": true,
        "type": "User",
        "userConsentDescription": "Allow the application to access the Dehydrated API on your behalf",
        "userConsentDisplayName": "Access Dehydrated API",
        "value": "access_as_user"
      }
    ]
  }
}
EOF
    
    # Add delegated permission to the API application
    az rest --method PATCH \
        --uri "https://graph.microsoft.com/v1.0/applications/$(az ad app show --id $API_APP_ID --query 'id' -o tsv)" \
        --headers "Content-Type=application/json" \
        --body @delegated-permission.json
    
    # Clean up temporary file
    rm delegated-permission.json
    
    print_success "Created delegated permission for API application"
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

# Function to configure SPA redirect URI for client application
configure_spa_redirect_uri() {
    print_status "Configuring SPA redirect URI for client application..."
    
    # Default development redirect URI
    DEFAULT_REDIRECT_URI="http://localhost:5173/auth-callback"
    
    # Configure the SPA redirect URI for the client application
    # Use REST API to set SPA redirect URIs (Azure CLI doesn't have --spa-redirect-uris)
    print_status "Setting SPA redirect URI using REST API..."
    az rest --method PATCH \
        --uri "https://graph.microsoft.com/v1.0/applications/$(az ad app show --id $CLIENT_APP_ID --query 'id' -o tsv)" \
        --headers "Content-Type=application/json" \
        --body "{\"spa\":{\"redirectUris\":[\"$DEFAULT_REDIRECT_URI\"]},\"web\":{\"redirectUris\":[]}}"
    
    print_success "Configured SPA redirect URI: $DEFAULT_REDIRECT_URI"
    print_status "For production, update the redirect URI using:"
    echo "  az rest --method PATCH --uri \"https://graph.microsoft.com/v1.0/applications/\$(az ad app show --id $CLIENT_APP_ID --query 'id' -o tsv)\" \\"
    echo "    --headers \"Content-Type=application/json\" \\"
    echo "    --body '{\"spa\":{\"redirectUris\":[\"https://your-domain.com/auth-callback\"]},\"web\":{\"redirectUris\":[]}}'"
    echo ""
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
    
    # Get the delegated permission ID from the API application
    ACCESS_AS_USER_PERMISSION_ID=$(az ad app show --id $API_APP_ID --query "api.oauth2PermissionScopes[?value=='access_as_user'].id" -o tsv)

    if [ -n "$ACCESS_AS_USER_PERMISSION_ID" ]; then
        print_status "Access as user permission ID: $ACCESS_AS_USER_PERMISSION_ID"
    fi
    
    # Check if permission already exists
    EXISTING_PERMISSION=$(az ad app permission list --id $CLIENT_APP_ID --query "[?resourceAppId=='$API_APP_ID'].resourceAppId" -o tsv)
    
    if [ "$EXISTING_PERMISSION" = "$API_APP_ID" ]; then
        print_warning "API permission already exists"
    else
        # Grant API delegated permission if it exists
        if [ -n "$ACCESS_AS_USER_PERMISSION_ID" ]; then
            az ad app permission add \
                --id $CLIENT_APP_ID \
                --api $API_APP_ID \
                --api-permissions "$ACCESS_AS_USER_PERMISSION_ID=Scope"
        fi
        
        print_success "Granted API permissions"
    fi
    
    # Grant admin consent with retry logic
    print_status "Granting admin consent..."
    
    # Add a delay before attempting admin consent to allow Azure AD to propagate changes
    print_status "Waiting for Azure AD to propagate permission changes..."
    sleep 10
    
    MAX_RETRIES=5
    RETRY_COUNT=0
    
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if az ad app permission admin-consent --id $CLIENT_APP_ID 2>/dev/null; then
            print_success "Admin consent granted"
            break
        else
            RETRY_COUNT=$((RETRY_COUNT + 1))
            if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
                print_warning "Admin consent failed, retrying in 10 seconds... (attempt $RETRY_COUNT/$MAX_RETRIES)"
                sleep 10
            else
                print_warning "Admin consent failed after $MAX_RETRIES attempts. You may need to grant consent manually."
                print_status "You can grant consent manually by visiting:"
                echo "  https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps/ApplicationMenuBlade/CallAnAPI/appId/$CLIENT_APP_ID"
                echo ""
                print_status "Or run this command manually:"
                echo "  az ad app permission admin-consent --id $CLIENT_APP_ID"
                echo ""
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
        
        # Check if delegated permissions are configured
        DELEGATED_PERMISSIONS_COUNT=$(az ad app show --id $API_APP_ID --query "api.oauth2PermissionScopes | length(@)" -o tsv)
        if [ "$DELEGATED_PERMISSIONS_COUNT" -gt 0 ]; then
            print_success "API application has $DELEGATED_PERMISSIONS_COUNT delegated permissions configured"
        else
            print_warning "API application has no delegated permissions configured"
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
    echo "  4. For frontend testing, use the delegated permission scope:"
    echo "     api://$API_APP_ID/access_as_user"
    echo ""
}

# Function to generate configuration file
generate_config() {
    print_status "Generating configuration file..."
    
    CONFIG_FILE="examples/config-auth.yaml"
    
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
    - "api://$API_APP_ID"
  # Enable JWT signature validation (recommended for production)
  enableSignatureValidation: true
  # Key cache TTL (e.g., "24h", "1h", "30m")
  keyCacheTTL: "24h"
EOF
    
    print_success "Configuration file generated: $CONFIG_FILE"
}

# Function to generate frontend configuration file
generate_frontend_config() {
    print_status "Generating frontend configuration file..."
    
    FRONTEND_CONFIG_FILE="frontend.env.example"
    
    cat > $FRONTEND_CONFIG_FILE << EOF
# Frontend Environment Configuration for Dehydrated API
# Copy this file to your frontend project as .env

# Enable MSAL authentication
VITE_ENABLE_MSAL=true

# Azure AD Client Application ID (for frontend authentication)
VITE_MSAL_CLIENT_ID=$CLIENT_APP_ID

# Azure AD Authority URL
VITE_MSAL_AUTHORITY=https://login.microsoftonline.com/$TENANT_ID

# Dehydrated API Identifier (used for token scopes)
VITE_DEHYDRATED_API_IDENTIFIER=api://$API_APP_ID

# Optional: API Base URL (update if different from default)
# VITE_API_BASE_URL=http://localhost:3000
EOF
    
    print_success "Frontend configuration file generated: $FRONTEND_CONFIG_FILE"
    print_status "Copy this file to your frontend project:"
    echo "  cp $FRONTEND_CONFIG_FILE ../dehydrated-frontend/.env"
    echo ""
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
    echo "     --resource api://$API_APP_ID \\"
    echo "     --test-api"
    echo ""
    echo "5. For development, you can also use:"
    echo "   ./scripts/get-custom-api-token.sh \\"
    echo "     --tenant-id $TENANT_ID \\"
    echo "     --client-id $CLIENT_APP_ID \\"
    echo "     --client-secret YOUR_SECRET \\"
    echo "     --resource api://$API_APP_ID \\"
    echo "     --output token"
    echo ""
    echo "Frontend Configuration:"
    echo "Add the following environment variables to your frontend .env file:"
    echo ""
    echo "  VITE_ENABLE_MSAL=true"
    echo "  VITE_MSAL_CLIENT_ID=$CLIENT_APP_ID"
    echo "  VITE_MSAL_AUTHORITY=https://login.microsoftonline.com/$TENANT_ID"
    echo "  VITE_DEHYDRATED_API_IDENTIFIER=api://$API_APP_ID"
    echo ""
    echo "A frontend configuration file has been generated: frontend.env.example"
    echo "Copy it to your frontend project:"
    echo "  cp frontend.env.example ../dehydrated-frontend/.env"
    echo ""
    echo "Or manually add these variables to your frontend .env file:"
    echo "  VITE_ENABLE_MSAL=true"
    echo "  VITE_MSAL_CLIENT_ID=$CLIENT_APP_ID"
    echo "  VITE_MSAL_AUTHORITY=https://login.microsoftonline.com/$TENANT_ID"
    echo "  VITE_DEHYDRATED_API_IDENTIFIER=api://$API_APP_ID"
    echo ""
    echo "SPA Redirect URI Configuration:"
    echo "The client application has been configured with the default development redirect URI:"
    echo "  http://localhost:5173/auth-callback"
    echo ""
    echo "For production deployment, update the redirect URI using:"
    echo "  az ad app update --id $CLIENT_APP_ID --web-redirect-uris https://your-domain.com/auth-callback"
    echo ""
    echo "For multiple environments, you can add multiple redirect URIs:"
    echo "  az ad app update --id $CLIENT_APP_ID --web-redirect-uris \\"
    echo "    http://localhost:5173/auth-callback \\"
    echo "    https://your-domain.com/auth-callback \\"
    echo "    https://staging.your-domain.com/auth-callback"
    echo ""
    echo "Delegated Permissions Available:"
    echo "  - access_as_user: For user authentication (delegated permission)"
    echo ""
    echo "Identifier URIs:"
    echo "  - api://$API_APP_ID"
    echo ""
    echo "Script Options:"
    echo "  --delete   Only delete existing applications (no creation)"
    echo "  --force    Delete existing applications and recreate them (no prompts)"
    echo "  --help     Show script usage information"
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
    echo "./scripts/azure-ad-setup.sh --delete"
    echo ""
}

# Main execution
main() {
    echo "=========================================="
    echo "Azure AD Setup for Dehydrated API Go"
    echo "Using Custom API"
    echo "=========================================="
    echo ""
    
    # Parse command line arguments
    parse_args "$@"
    
    # Show mode status
    if [ "$DELETE_ONLY" = true ] && [ "$FORCE_RECREATE" = true ]; then
        print_warning "Force delete mode enabled - existing applications will be deleted without prompts (no creation)"
        echo ""
    elif [ "$DELETE_ONLY" = true ]; then
        print_warning "Delete mode enabled - existing applications will be deleted with prompts (no creation)"
        echo ""
    elif [ "$FORCE_RECREATE" = true ]; then
        print_warning "Force mode enabled - existing applications will be deleted and recreated without prompts"
        echo ""
    else
        print_status "Interactive mode - you will be prompted for existing application deletion"
        echo ""
    fi
    
    check_azure_cli
    get_tenant_info
    cleanup_existing_apps
    
    # If in delete-only mode, exit after cleanup
    if [ "$DELETE_ONLY" = true ]; then
        print_success "Delete operation completed"
        echo ""
        echo "=========================================="
        echo "Delete Mode Complete"
        echo "=========================================="
        echo ""
        echo "Existing applications have been deleted."
        echo "To recreate applications, run the script without --delete"
        echo ""
        exit 0
    fi
    
    # Continue with creation if not in delete-only mode
    create_api_app
    create_delegated_permission
    create_client_app
    configure_spa_redirect_uri
    create_service_principals
    grant_api_permissions
    test_token_generation
    generate_config
    generate_frontend_config
    display_instructions
    display_cleanup_instructions
}

# Run main function
main "$@" 