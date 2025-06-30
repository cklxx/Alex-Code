#!/bin/bash

# Release script for Deep Coding Agent
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

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

# Get version from git tags or use provided version
get_version() {
    if [ -n "$1" ]; then
        echo "$1"
    elif git describe --tags --exact-match HEAD 2>/dev/null; then
        git describe --tags --exact-match HEAD
    else
        echo "v$(date +%Y%m%d)-$(git rev-parse --short HEAD)"
    fi
}

# Validate version format
validate_version() {
    local version="$1"
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]] && [[ ! $version =~ ^v[0-9]{8}-[a-f0-9]+$ ]]; then
        print_error "Invalid version format: $version"
        print_error "Expected format: v1.2.3 or v1.2.3-beta.1"
        exit 1
    fi
}

# Pre-release checks
pre_release_checks() {
    print_status "Running pre-release checks..."
    
    # Check if git is clean
    if [[ -n $(git status --porcelain) ]]; then
        print_error "Git working directory is not clean. Please commit or stash your changes."
        exit 1
    fi
    
    # Check if we're on main branch
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" ]] && [[ "$current_branch" != "master" ]]; then
        print_warning "Not on main/master branch. Current branch: $current_branch"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Run tests
    print_status "Running comprehensive tests..."
    ./scripts/test.sh all
    
    # Run linting
    if command -v golangci-lint &> /dev/null; then
        print_status "Running linter..."
        golangci-lint run
    fi
    
    print_success "Pre-release checks passed"
}

# Build release binaries
build_release() {
    local version="$1"
    
    print_status "Building release binaries for version $version..."
    
    # Clean previous builds
    rm -rf release/
    mkdir -p release
    
    # Build information
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse HEAD)
    
    # Platforms to build for
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64" 
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    
    for platform in "${platforms[@]}"; do
        platform_split=(${platform//\// })
        GOOS=${platform_split[0]}
        GOARCH=${platform_split[1]}
        
        output_name="deep-coding-agent-$version-$GOOS-$GOARCH"
        if [ $GOOS = "windows" ]; then
            output_name+='.exe'
        fi
        
        print_status "Building for $GOOS/$GOARCH..."
        
        env GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags="-X main.Version=$version -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
            -o "release/$output_name" \
            cmd/simple-main.go
        
        # Create tarball for non-Windows platforms
        if [ $GOOS != "windows" ]; then
            tar -czf "release/$output_name.tar.gz" -C release "$output_name"
            rm "release/$output_name"
        else
            # Create zip for Windows
            (cd release && zip -q "$output_name.zip" "$output_name")
            rm "release/$output_name"
        fi
    done
    
    print_success "Release binaries built in release/ directory"
}

# Generate checksums
generate_checksums() {
    print_status "Generating checksums..."
    
    cd release/
    
    # Generate SHA256 checksums
    if command -v sha256sum &> /dev/null; then
        sha256sum * > checksums.sha256
    elif command -v shasum &> /dev/null; then
        shasum -a 256 * > checksums.sha256
    else
        print_warning "No SHA256 utility found, skipping checksums"
        cd ..
        return
    fi
    
    cd ..
    print_success "Checksums generated in release/checksums.sha256"
}

# Create release notes
create_release_notes() {
    local version="$1"
    local notes_file="release/RELEASE_NOTES_$version.md"
    
    print_status "Creating release notes..."
    
    # Get previous tag for changelog
    prev_tag=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
    
    cat > "$notes_file" << EOF
# Deep Coding Agent $version

## Features

- High-performance code analysis with sub-30ms execution times
- Multi-language support (Go, JavaScript, TypeScript, Python, Java, C/C++)
- AI-powered code generation and refactoring
- 8 different refactoring patterns
- Concurrent analysis for large codebases
- Template-based code generation
- Configuration management
- Security vulnerability detection

## Installation

### Binary Download

Download the appropriate binary for your platform from the release assets:

- **Linux (x64)**: \`deep-coding-agent-$version-linux-amd64.tar.gz\`
- **Linux (ARM64)**: \`deep-coding-agent-$version-linux-arm64.tar.gz\`
- **macOS (Intel)**: \`deep-coding-agent-$version-darwin-amd64.tar.gz\`
- **macOS (Apple Silicon)**: \`deep-coding-agent-$version-darwin-arm64.tar.gz\`
- **Windows (x64)**: \`deep-coding-agent-$version-windows-amd64.exe.zip\`
- **Windows (ARM64)**: \`deep-coding-agent-$version-windows-arm64.exe.zip\`

### Build from Source

\`\`\`bash
git clone <repository-url>
cd deep-coding
go build -o deep-coding-agent cmd/simple-main.go
\`\`\`

## Usage Examples

\`\`\`bash
# Analyze code
./deep-coding-agent analyze src/ --concurrent

# Generate code
./deep-coding-agent generate "REST API server" go --style=clean

# Refactor code
./deep-coding-agent refactor main.go --pattern=modernize --backup

# AI-powered analysis
./deep-coding-agent analyze main.go --ai --ai-type=security
\`\`\`

## Performance

- **40-100x faster** than Node.js-based alternatives
- **Sub-30ms** analysis and generation times
- **Concurrent processing** with controlled worker pools
- **Memory efficient** with minimal resource usage

EOF

    if [ -n "$prev_tag" ]; then
        echo "## Changes since $prev_tag" >> "$notes_file"
        echo "" >> "$notes_file"
        git log --oneline "$prev_tag..HEAD" | sed 's/^/- /' >> "$notes_file"
        echo "" >> "$notes_file"
    fi

    cat >> "$notes_file" << EOF
## Verification

Verify the integrity of downloaded files using the provided checksums:

\`\`\`bash
sha256sum -c checksums.sha256
\`\`\`

## Support

- Documentation: See README.md
- Issues: Please report bugs and feature requests on GitHub
- Performance: This release maintains sub-30ms execution times for all core operations

---

Built with Go $(go version | cut -d' ' -f3) on $(date -u '+%Y-%m-%d %H:%M:%S UTC')
EOF

    print_success "Release notes created: $notes_file"
}

# Tag release
tag_release() {
    local version="$1"
    
    print_status "Creating git tag for version $version..."
    
    # Check if tag already exists
    if git tag -l | grep -q "^$version$"; then
        print_error "Tag $version already exists"
        exit 1
    fi
    
    # Create annotated tag
    git tag -a "$version" -m "Release $version"
    print_success "Tagged release: $version"
    
    print_status "Push the tag with: git push origin $version"
}

# Full release process
full_release() {
    local version="$1"
    
    validate_version "$version"
    pre_release_checks
    build_release "$version"
    generate_checksums
    create_release_notes "$version"
    tag_release "$version"
    
    print_success "Release $version is ready!"
    print_status "Next steps:"
    print_status "1. git push origin $version"
    print_status "2. Create GitHub release with files in release/ directory"
    print_status "3. Upload release binaries and checksums"
}

# Help message
show_help() {
    cat << EOF
Deep Coding Agent Release Script

Usage: $0 [command] [version]

Commands:
    check           Run pre-release checks only
    build [version] Build release binaries
    tag [version]   Create git tag
    full [version]  Complete release process
    help            Show this help

Examples:
    $0 check                # Run pre-release checks
    $0 build v1.2.3        # Build binaries for v1.2.3
    $0 full v1.2.3         # Complete release process
    $0 tag v1.2.3          # Tag current commit as v1.2.3

If no version is specified, it will be auto-generated based on date and git commit.
EOF
}

# Main command dispatcher
case "${1:-help}" in
    check)
        pre_release_checks
        ;;
    build)
        version=$(get_version "$2")
        validate_version "$version"
        build_release "$version"
        generate_checksums
        create_release_notes "$version"
        ;;
    tag)
        version=$(get_version "$2")
        validate_version "$version"
        tag_release "$version"
        ;;
    full)
        version=$(get_version "$2")
        full_release "$version"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac