# 🎯 GHE - Beautiful GitHub Account Switcher (Go Version)

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

*✨ A beautiful, interactive CLI tool for seamlessly managing multiple GitHub accounts per repository*

This is the Go implementation of GHE, migrated from the original TypeScript/Bun version.

## 🚀 Quick Start

```bash
# Start interactive mode
ghe

# Clone repository with account selection
ghe https://github.com/user/repo.git
ghe git@github.com:user/repo.git

# Check version
ghe version

# Get help
ghe help
```

## 📦 Installation

### From Source

```bash
# Clone and build
git clone https://github.com/dwirx/ghex.git
cd ghex
make build

# Install to PATH
sudo mv build/ghe /usr/local/bin/
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/dwirx/ghex/releases):

- `ghe-linux-amd64` - Linux x64
- `ghe-linux-arm64` - Linux ARM64
- `ghe-darwin-amd64` - macOS Intel
- `ghe-darwin-arm64` - macOS Apple Silicon
- `ghe-windows-amd64.exe` - Windows x64

## 🌟 Features

- 🎨 **Beautiful Terminal UI** - Colorful and intuitive interface
- 🔄 **Multi-Account Support** - Switch between different GitHub accounts
- 🔐 **Dual Authentication** - SSH keys and Personal Access Tokens
- 📁 **Per-Repository Config** - Different accounts for different repos
- 📦 **Git Clone Integration** - Clone with account selection
- ⚡ **Single Binary** - No runtime dependencies
- 🖥️ **Cross-Platform** - Windows, Linux, macOS support

## 🛠️ Commands

### Interactive Mode
```bash
ghe              # Start interactive menu
```

### Account Management
```bash
ghe list         # List all accounts
ghe status       # Show current repo status
ghe switch       # Switch account for current repo
ghe switch work  # Switch to specific account
ghe health       # Check health of all accounts
ghe log          # View activity log
```

### Git Shortcuts
```bash
ghe gs           # git status
ghe gb           # git branch
ghe gba          # git branch -a
ghe gf           # git fetch origin
ghe gp           # git pull
ghe gco main     # git checkout main
ghe shove "msg"  # git add, commit, push
```

### Git Config
```bash
ghe setname "John Doe"      # Set global user.name
ghe setmail john@email.com  # Set global user.email
ghe showconfig              # Show git config
```

## 🔧 Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Original TypeScript version: [ghe](https://github.com/dwirx/ghe)
- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- UI powered by [Charm](https://charm.sh) libraries
