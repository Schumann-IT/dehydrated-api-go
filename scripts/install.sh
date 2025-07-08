#!/bin/bash

# Dehydrated API Go Install Script
# Downloads and installs a specific version of dehydrated-api-go

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
REPO="schumann-it/dehydrated-api-go"
BINARY_NAME="dehydrated-api-go"
TEMP_DIR="/tmp/dehydrated-api-go-install"

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

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 <version>

Downloads and installs dehydrated-api-go binary for the specified version to the current directory.

Arguments:
  version     The version to install (e.g., v1.0.0, 1.0.0)

Examples:
  $0 v1.0.0
  $0 1.0.0

Note: Version is mandatory as GitHub releases do not support 'latest' version.
The binary will be installed to the current working directory.
EOF
}

# Function to detect OS and architecture
detect_platform() {
    local os
    local arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="Linux" ;;
        Darwin*)    os="Darwin" ;;
        CYGWIN*)    os="Windows" ;;
        MINGW*)     os="Windows" ;;
        MSYS*)      os="Windows" ;;
        *)          os="$(uname -s)" ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64)     arch="x86_64" ;;
        amd64)      arch="x86_64" ;;
        arm64)      arch="arm64" ;;
        aarch64)    arch="arm64" ;;
        i386)       arch="i386" ;;
        i686)       arch="i386" ;;
        *)          arch="$(uname -m)" ;;
    esac
    
    echo "${os}_${arch}"
}

# Function to check if binary exists and is executable
check_binary() {
    if [[ -x "./$BINARY_NAME" ]]; then
        return 0
    else
        return 1
    fi
}

# Function to get current version
get_current_version() {
    if check_binary; then
        local version_output
        version_output=$("./$BINARY_NAME" -version 2>/dev/null | head -n1)
        local version
        version=$(echo "$version_output" | grep -o 'version [^ ]*' | cut -d' ' -f2)
        if [[ -z "$version" ]]; then
            # fallback: print raw output for debugging
            echo "unknown (raw output: $version_output)"
        else
            echo "$version"
        fi
    else
        echo "not installed"
    fi
}

# Function to download and install
install_version() {
    local version="$1"
    local platform="$2"
    local original_dir="$(pwd)"
    
    # Create temporary directory
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # Construct download URL
    local archive_name="${BINARY_NAME}_${platform}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    
    print_status "Downloading ${BINARY_NAME} version ${version} for ${platform}..."
    print_status "URL: ${download_url}"
    
    # Download the archive
    if ! curl -L -f -o "${archive_name}" "$download_url"; then
        print_error "Failed to download ${archive_name}"
        print_error "Please check if version ${version} exists and supports your platform"
        print_error "Available platforms: Linux_x86_64, Linux_arm64, Darwin_x86_64, Darwin_arm64"
        return 1
    fi
    
    print_status "Extracting archive..."
    tar -xzf "${archive_name}"
    
    # Check if binary was extracted
    if [[ ! -f "$BINARY_NAME" ]]; then
        print_error "Binary not found in extracted archive"
        return 1
    fi
    
    # Make binary executable
    chmod +x "$BINARY_NAME"
    
    # Move binary to install directory (original directory)
    mv "$BINARY_NAME" "$original_dir/"
    
    print_success "Installed ${BINARY_NAME} version ${version} to current directory"
    
    # Clean up
    cd "$original_dir"
    rm -rf "$TEMP_DIR"
}

# Function to verify installation
verify_installation() {
    local version="$1"
    
    if check_binary; then
        local installed_version
        installed_version=$(get_current_version)

        if [[ "$installed_version" == "unknown"* ]]; then
            print_warning "Installed binary doesn't report version, but binary is available"
            print_success "Installation completed successfully"
        elif [[ "$installed_version" == "$version" ]]; then
            print_success "Installation verified: ${BINARY_NAME} version ${installed_version} is now available in current directory"
        else
            print_warning "Version mismatch: expected ${version}, got ${installed_version}"
            print_success "Installation completed, but version verification failed"
        fi
    else
        print_error "Installation failed: ${BINARY_NAME} is not available in current directory"
        return 1
    fi
}

# Main script logic
main() {
    # Check if version is provided
    if [[ $# -eq 0 ]]; then
        print_error "Version is required"
        show_usage
        exit 1
    fi
    
    local version="$1"
    
    # Remove 'v' prefix if present for consistency
    if [[ "$version" =~ ^v[0-9] ]]; then
        version="${version#v}"
    fi
    
    # Add 'v' prefix for GitHub releases
    local github_version="v${version}"
    
    print_status "Installing ${BINARY_NAME} version ${version}"
    
    # Check current installation
    local current_version
    current_version=$(get_current_version)
    if [[ "$current_version" != "not installed" ]]; then
        print_warning "Current version: ${current_version}"
        read -p "Do you want to continue with the installation? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Installation cancelled"
            exit 0
        fi
    fi
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    print_status "Detected platform: ${platform}"
    
    # Check if platform is supported
    case "$platform" in
        Linux_x86_64|Linux_arm64|Darwin_x86_64|Darwin_arm64)
            ;;
        *)
            print_error "Unsupported platform: ${platform}"
            print_error "Supported platforms: Linux_x86_64, Linux_arm64, Darwin_x86_64, Darwin_arm64"
            exit 1
            ;;
    esac
    
    # Install the version
    if install_version "$github_version" "$platform"; then
        verify_installation "$version"
    else
        print_error "Installation failed"
        exit 1
    fi
}

# Run main function with all arguments
main "$@" 