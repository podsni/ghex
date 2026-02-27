# GHEX Installer Script for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/dwirx/ghex/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "dwirx/ghex"
$BinaryName = "ghex.exe"
$InstallDir = "$env:LOCALAPPDATA\ghex"

function Write-Banner {
    Write-Host @"
  ██████╗ ██╗  ██╗███████╗██╗  ██╗
 ██╔════╝ ██║  ██║██╔════╝╚██╗██╔╝
 ██║  ███╗███████║█████╗   ╚███╔╝ 
 ██║   ██║██╔══██║██╔══╝   ██╔██╗ 
 ╚██████╔╝██║  ██║███████╗██╔╝ ██╗
  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
"@ -ForegroundColor Cyan
    Write-Host "GitHub Account Switcher & Universal Downloader" -ForegroundColor White
    Write-Host ""
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -ForegroundColor Blue -NoNewline
    Write-Host $Message
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-ErrorMsg {
    param([string]$Message)
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
    throw $Message
}

function Get-Architecture {
    $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-ErrorMsg "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $response.tag_name
    }
    catch {
        Write-ErrorMsg "Failed to get latest version: $_"
    }
}

function Install-Ghex {
    param(
        [string]$Version,
        [string]$Arch
    )
    
    $filename = "ghex-windows-$Arch.zip"
    $url = "https://github.com/$Repo/releases/download/$Version/$filename"
    
    Write-Info "Downloading $filename..."
    
    $tempDir = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "ghex-install-$(Get-Random)")
    $zipPath = Join-Path $tempDir $filename
    
    try {
        Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing
        
        Write-Info "Extracting..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
        
        # Binary name from goreleaser is just "ghex.exe"
        $binary = Join-Path $tempDir "ghex.exe"
        
        # Debug: show extracted files
        Write-Info "Extracted files:"
        Get-ChildItem $tempDir | ForEach-Object { Write-Host "  $_" }
        
        if (-not (Test-Path $binary)) {
            Write-ErrorMsg "Binary 'ghex.exe' not found after extraction"
        }
        
        Write-Info "Installing to $InstallDir..."
        
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        $destPath = Join-Path $InstallDir $BinaryName
        Copy-Item -Path $binary -Destination $destPath -Force
        
        # Add to PATH if not already there
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($currentPath -notlike "*$InstallDir*") {
            Write-Info "Adding to PATH..."
            [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
            $env:Path = "$env:Path;$InstallDir"
        }
    }
    finally {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Test-Installation {
    $ghexPath = Join-Path $InstallDir $BinaryName
    if (Test-Path $ghexPath) {
        Write-Success "GHEX installed successfully!"
        Write-Host ""
        Write-Host "  Location: $ghexPath" -ForegroundColor White
        Write-Host ""
        Write-Host "Please restart your terminal to use 'ghex' command." -ForegroundColor Yellow
        Write-Host "Or run: " -NoNewline
        Write-Host "$ghexPath --help" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "To update GHEX later, run: ghex update" -ForegroundColor Green
    }
    else {
        Write-Warn "Installation completed but binary not found at expected location"
    }
}

function Main {
    param([string]$Version)
    
    Write-Banner
    
    $arch = Get-Architecture
    Write-Info "Detected Architecture: $arch"
    
    if (-not $Version) {
        $Version = Get-LatestVersion
    }
    
    if (-not $Version) {
        Write-ErrorMsg "Could not determine latest version"
    }
    
    Write-Info "Installing GHEX $Version..."
    
    Install-Ghex -Version $Version -Arch $arch
    Test-Installation
}

# Run main
try {
    Main -Version $args[0]
} catch {
    Write-Host "[ERROR] $_" -ForegroundColor Red
    exit 1
}
