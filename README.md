# 🎯 GHEX - Beautiful GitHub Account Switcher & Universal Downloader

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

*✨ A beautiful, interactive CLI tool for seamlessly managing multiple GitHub accounts per repository with universal download capabilities*

## 🚀 Quick Start

```bash
# Start interactive mode
ghex

# Clone repository with account selection
ghex https://github.com/user/repo.git

# Download any file
ghex dlx https://example.com/file.zip

# Check version
ghex version
```

## 📦 Installation

### From Source

```bash
git clone https://github.com/dwirx/ghex.git
cd ghex
make build
sudo make install
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/dwirx/ghex/releases):

- `ghex-linux-amd64` - Linux x64
- `ghex-linux-arm64` - Linux ARM64
- `ghex-darwin-amd64` - macOS Intel
- `ghex-darwin-arm64` - macOS Apple Silicon
- `ghex-windows-amd64.exe` - Windows x64

## 🌟 Features

### Account Management
- 🔄 **Multi-Account Support** - Switch between different GitHub accounts
- 🔐 **Dual Authentication** - SSH keys and Personal Access Tokens
- 📁 **Per-Repository Config** - Different accounts for different repos
- 📦 **Git Clone Integration** - Clone with account selection
- 🏥 **Health Check** - Verify all account connections

### Universal Downloader (dlx)
- 📥 **Any URL Download** - Download files from any HTTP/HTTPS URL
- 📄 **Git File Download** - Download single files from GitHub/GitLab
- 📁 **Git Directory Download** - Download entire directories
- 🏷️ **Release Download** - Download GitHub release assets
- 📋 **Batch Download** - Download from URL list file

### Other Features
- 🎨 **Beautiful Terminal UI** - Colorful and intuitive interface
- ⚡ **Single Binary** - No runtime dependencies
- 🖥️ **Cross-Platform** - Windows, Linux, macOS support

## 🛠️ Commands

### Interactive Mode
```bash
ghex              # Start interactive menu
```

### Account Management
```bash
ghex list         # List all accounts
ghex status       # Show current repo status
ghex switch       # Switch account for current repo
ghex switch work  # Switch to specific account
ghex add          # Add new account
ghex edit         # Edit account
ghex remove       # Remove account
ghex health       # Check health of all accounts
ghex log          # View activity log
```

### SSH Management
```bash
ghex ssh              # SSH management menu
ghex ssh generate     # Generate new SSH key
ghex ssh import       # Import existing SSH key
ghex ssh test         # Test SSH connection
ghex ssh global       # Switch SSH globally
ghex ssh list         # List SSH keys
```

### Download (dlx)
```bash
# Download any file
ghex dlx https://example.com/file.zip
ghex dlx -o myfile.zip https://example.com/file.zip
ghex dlx -d ./downloads https://example.com/file.zip

# Download from Git repository
ghex dlx file https://github.com/user/repo/blob/main/README.md
ghex dlx dir https://github.com/user/repo/tree/main/src
ghex dlx release https://github.com/user/repo

# Download from URL list
ghex dlx list urls.txt
```

### Git Shortcuts
```bash
ghex gs           # git status
ghex gb           # git branch
ghex gba          # git branch -a
ghex gbr          # git branch -r
ghex gf           # git fetch origin
ghex gp           # git pull
ghex gpr          # git pull --rebase
ghex gco main     # git checkout main
ghex gcb feature  # git checkout -b feature
ghex gl           # git log --oneline
ghex gd           # git diff
ghex gds          # git diff --staged
ghex gst          # git stash
ghex gstp         # git stash pop
ghex greset       # git reset HEAD
ghex shove "msg"  # git add, commit, push
ghex shovenc "msg"# git add, commit, push (no confirm)
```

### Git Config
```bash
ghex setname "John Doe"      # Set global user.name
ghex setmail john@email.com  # Set global user.email
ghex showconfig              # Show git config
```

## 🔧 Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Install to /usr/local/bin
sudo make install

# Clean build artifacts
make clean
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- UI powered by [Charm](https://charm.sh) libraries (lipgloss, bubbletea)
