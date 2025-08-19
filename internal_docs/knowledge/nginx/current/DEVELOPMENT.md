# Development Guide

## 🎯 Overview

This guide covers development practices, setup, and workflow for the Falco nginx plugin project.

## 🔧 Development Setup

### Prerequisites

- Go 1.22+
- Make
- Docker (for testing)
- Falco 0.36.0+ (for runtime testing)

### Local Development

```bash
# Clone repository
git clone https://github.com/takaosgb3/falco-nginx-plugin-claude.git
cd falco-nginx-plugin-claude

# Install dependencies
go mod download

# Build SDK plugin
make build-sdk

# Run tests
make test

# Check code quality
make lint
```

## 📝 Code Organization

### Project Structure

```
.
├── cmd/
│   ├── plugin-sdk/     # SDK-based plugin (current)
│   └── test-runner/    # Test utilities
├── pkg/
│   ├── parser/         # Log parsing
│   ├── plugin/         # Plugin implementation
│   └── watcher/        # File watching
└── scripts/            # Automation scripts
```

### Key Components

1. **Plugin SDK Implementation** (`cmd/plugin-sdk/nginx.go`)
   - Uses Falco Plugin SDK for Go v0.8.1
   - Implements source and extractor interfaces
   - GOB encoding for event serialization

2. **Log Parser** (`pkg/parser/`)
   - Parses nginx access logs
   - Supports combined and custom formats
   - Extracts 17+ fields

3. **File Watcher** (`pkg/watcher/`)
   - Monitors log file changes
   - Handles log rotation
   - Efficient tail implementation

## 🔄 Development Workflow

### Branch Strategy

```bash
# Feature development
git checkout -b feature/your-feature

# Bug fixes
git checkout -b fix/issue-description

# Documentation
git checkout -b docs/update-description
```

### Testing

```bash
# Unit tests
go test ./pkg/...

# Integration tests
make test-integration

# Coverage report
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Lint check
make lint

# Security scan
make security-scan
```

## 🚀 Build and Release

### Building the Plugin

```bash
# Linux build (production)
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -buildmode=c-shared \
  -o libfalco-nginx-plugin.so \
  ./cmd/plugin-sdk

# Local testing build
make build-sdk
```

### Release Process

⚠️ **Important**: Always use GitHub Actions for releases

```bash
# Create release via workflow
gh workflow run build-and-release.yml \
  -f version=v1.2.XX \
  --repo takaosgb3/falco-nginx-plugin-claude
```

## 🔍 Debugging

### Common Issues

1. **Plugin not loading**
   ```bash
   # Check plugin in standalone mode
   sudo falco -c /etc/falco/falco.yaml \
     --disable-source syscall
   ```

2. **Rule not firing**
   - Verify `source: nginx` in rules
   - Check field names match exactly
   - Review log format compatibility

3. **Performance issues**
   - Monitor file watcher goroutines
   - Check log rotation handling
   - Review buffer sizes

### Debug Commands

```bash
# List loaded plugins
sudo falco --list-plugins

# Validate rules
sudo falco -V -r nginx_rules.yaml

# Test with verbose output
sudo falco -v --disable-source syscall
```

## 📚 Development Resources

### Documentation

- [Architecture Overview](./ARCHITECTURE.md)
- [Testing Guide](./TESTING.md)
- [CI/CD Pipeline](./CI_CD.md)
- [Security Guidelines](./SECURITY.md)

### External Resources

- [Falco Plugin SDK Documentation](https://github.com/falcosecurity/plugin-sdk-go)
- [Falco Rules Reference](https://falco.org/docs/rules/)
- [Go Best Practices](https://go.dev/doc/effective_go)

## 📈 Recent Updates

### 2025-08-14
- Documentation reorganization
- Simplified directory structure
- Updated development guides

### 2025-08-04
- Complete migration to SDK-based implementation
- Removed CGO dependencies
- Improved stability and performance

## 👥 Contributing

Contributions are welcome! Please:

1. Follow the code style guidelines
2. Add tests for new features
3. Update documentation as needed
4. Ensure CI checks pass

For questions or support, open an issue on GitHub.
