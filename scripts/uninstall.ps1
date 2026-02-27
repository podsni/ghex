# GHEX Uninstaller Script for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/dwirx/ghex/main/scripts/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$BinaryName = "ghex.exe"
$InstallDir = "$env:LOCALAPPDATA\ghex"
$ConfigDirPrimary = "$env:APPDATA\ghe"
$ConfigDirLegacy = "$env:APPDATA\github-switch"

function Write-Banner {
    Write-Host @"
  ██████╗ ██╗  ██╗███████╗██╗  ██╗
 ██╔════╝ ██║  ██║██╔════╝╚██╗██╔╝
 ██║  ███╗███████║█████╗   ╚███╔╝ 
 ██║   ██║██╔══██║██╔══╝   ██╔██╗ 
 ╚██████╔╝██║  ██║███████╗██╔╝ ██╗
  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
"@ -ForegroundColor Red
    Write-Host "GHEX Uninstaller" -ForegroundColor White
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

function Confirm-Action {
    param(
        [string]$Prompt,
        [bool]$Default = $false
    )
    
    if ($Default) {
        $promptText = "$Prompt [Y/n]: "
    } else {
        $promptText = "$Prompt [y/N]: "
    }
    
    $response = Read-Host $promptText
    
    if ([string]::IsNullOrWhiteSpace($response)) {
        return $Default
    }
    
    return $response -match '^[yY]'
}

function Remove-GhexBinary {
    $binaryPath = Join-Path $InstallDir $BinaryName
    
    if (Test-Path $binaryPath) {
        Write-Info "Removing binary: $binaryPath"
        try {
            Remove-Item -Path $binaryPath -Force
            Write-Success "Binary removed"
        }
        catch {
            Write-ErrorMsg "Failed to remove binary: $_"
            Write-Host "Try running as Administrator or manually delete: $binaryPath" -ForegroundColor Yellow
            return $false
        }
    } else {
        Write-Warn "Binary not found at $binaryPath"
    }
    
    # Remove install directory if empty
    if ((Test-Path $InstallDir) -and ((Get-ChildItem $InstallDir | Measure-Object).Count -eq 0)) {
        Write-Info "Removing empty install directory: $InstallDir"
        Remove-Item -Path $InstallDir -Force
    }
    
    return $true
}

function Remove-FromPath {
    Write-Info "Removing from PATH..."
    
    try {
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        
        if ($currentPath -like "*$InstallDir*") {
            $newPath = ($currentPath -split ';' | Where-Object { $_ -ne $InstallDir }) -join ';'
            [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
            Write-Success "Removed from PATH"
        } else {
            Write-Info "Install directory not found in PATH"
        }
    }
    catch {
        Write-ErrorMsg "Failed to update PATH: $_"
        Write-Host "You may need to manually remove '$InstallDir' from your PATH environment variable" -ForegroundColor Yellow
    }
}

function Remove-GhexConfig {
    $removed = $false
    
    if (Test-Path $ConfigDirPrimary) {
        Write-Info "Removing config directory: $ConfigDirPrimary"
        try {
            Remove-Item -Path $ConfigDirPrimary -Recurse -Force
            Write-Success "Config directory removed: $ConfigDirPrimary"
            $removed = $true
        }
        catch {
            Write-ErrorMsg "Failed to remove config directory: $_"
        }
    }
    
    if (Test-Path $ConfigDirLegacy) {
        Write-Info "Removing legacy config directory: $ConfigDirLegacy"
        try {
            Remove-Item -Path $ConfigDirLegacy -Recurse -Force
            Write-Success "Legacy config directory removed: $ConfigDirLegacy"
            $removed = $true
        }
        catch {
            Write-ErrorMsg "Failed to remove legacy config directory: $_"
        }
    }
    
    if (-not $removed) {
        Write-Warn "No config directories found"
    }
}

function Show-Preview {
    Write-Host ""
    Write-Info "The following will be removed:"
    Write-Host ""
    
    $binaryPath = Join-Path $InstallDir $BinaryName
    if (Test-Path $binaryPath) {
        Write-Host "  Binary: $binaryPath" -ForegroundColor White
    } else {
        Write-Host "  Binary: (not found)" -ForegroundColor Gray
    }
    
    if (Test-Path $InstallDir) {
        Write-Host "  Install Dir: $InstallDir" -ForegroundColor White
    }
    
    if (Test-Path $ConfigDirPrimary) {
        Write-Host "  Config: $ConfigDirPrimary" -ForegroundColor White
    }
    
    if (Test-Path $ConfigDirLegacy) {
        Write-Host "  Legacy Config: $ConfigDirLegacy" -ForegroundColor White
    }
    
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -like "*$InstallDir*") {
        Write-Host "  PATH entry: $InstallDir" -ForegroundColor White
    }
    
    Write-Host ""
}

function Main {
    param(
        [switch]$Purge,
        [switch]$Force
    )
    
    Write-Banner
    Show-Preview
    
    # Confirm uninstallation
    if (-not $Force) {
        if (-not (Confirm-Action "Do you want to uninstall GHEX?")) {
            Write-Info "Uninstallation cancelled"
            return
        }
    }
    
    # Remove binary
    $binaryRemoved = Remove-GhexBinary
    if (-not $binaryRemoved) {
        Write-Warn "Binary removal failed. Continuing with other cleanup..."
    }
    
    # Remove from PATH
    Remove-FromPath
    
    # Handle config removal
    if ($Purge) {
        Remove-GhexConfig
    } elseif (-not $Force) {
        Write-Host ""
        if (Confirm-Action "Do you want to remove configuration files as well?") {
            Remove-GhexConfig
        } else {
            Write-Info "Configuration files preserved"
        }
    }
    
    Write-Host ""
    Write-Success "GHEX has been uninstalled!"
    Write-Host ""
    Write-Host "Thank you for using GHEX! " -NoNewline
    Write-Host ([char]0x1F44B) # Wave emoji
    Write-Host ""
    Write-Host "Please restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
}

# Parse arguments and run
$purgeFlag = $args -contains "--purge" -or $args -contains "-p"
$forceFlag = $args -contains "--force" -or $args -contains "-f"

if ($args -contains "--help" -or $args -contains "-h") {
    Write-Host "Usage: uninstall.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  --purge, -p    Remove config files as well"
    Write-Host "  --force, -f    Skip confirmation prompts"
    Write-Host "  --help, -h     Show this help message"
    exit 0
}

try {
    Main -Purge:$purgeFlag -Force:$forceFlag
} catch {
    Write-Host "[ERROR] $_" -ForegroundColor Red
    exit 1
}
