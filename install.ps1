# install.ps1 — Install orchestratr on Windows
#
# Usage:
#   .\install.ps1
#
# This script:
#   1. Checks prerequisites (Go toolchain)
#   2. Builds orchestratr from source (or uses pre-built binary if available)
#   3. Installs the binary to %LOCALAPPDATA%\orchestratr (or $env:INSTALL_DIR)
#   4. Runs 'orchestratr install' to configure autostart
#
# Environment variables:
#   INSTALL_DIR — override install directory
#   SKIP_BUILD  — set to 1 to skip building

$ErrorActionPreference = "Stop"

$BinaryName = "orchestratr.exe"
$DefaultInstallDir = Join-Path $env:LOCALAPPDATA "orchestratr"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { $DefaultInstallDir }

function Write-Info($msg)  { Write-Host "  → $msg" -ForegroundColor Cyan }
function Write-Warn($msg)  { Write-Host "  ⚠ $msg" -ForegroundColor Yellow }
function Write-Err($msg)   { Write-Host "  ✗ $msg" -ForegroundColor Red }
function Write-Ok($msg)    { Write-Host "  ✓ $msg" -ForegroundColor Green }

function Test-Prerequisites {
    if ($env:SKIP_BUILD -eq "1") {
        $existing = Get-Command $BinaryName -ErrorAction SilentlyContinue
        if (-not $existing) {
            Write-Err "SKIP_BUILD=1 but '$BinaryName' not found in PATH"
            exit 1
        }
        Write-Info "Using existing $BinaryName from PATH"
        return
    }

    $go = Get-Command "go" -ErrorAction SilentlyContinue
    if (-not $go) {
        Write-Err "Go toolchain not found. Install Go from https://go.dev/dl/"
        exit 1
    }

    $goVersion = (go version) -replace '.*go(\d+\.\d+).*', '$1'
    Write-Info "Found Go $goVersion"
}

function Install-Binary {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

    if ($env:SKIP_BUILD -eq "1") {
        $binPath = (Get-Command $BinaryName).Source
        Write-Ok "Using $binPath"
        return
    }

    $scriptDir = $PSScriptRoot
    $goMod = Join-Path $scriptDir "go.mod"

    if (Test-Path $goMod) {
        Write-Info "Building from source in $scriptDir"
        $outputPath = Join-Path $InstallDir $BinaryName
        Push-Location $scriptDir
        try {
            go build -o $outputPath ./cmd/orchestratr
        } finally {
            Pop-Location
        }
    } else {
        Write-Info "Building via go install"
        $env:GOBIN = $InstallDir
        go install "github.com/josiahH-cf/orchestratr/cmd/orchestratr@latest"
    }

    $binPath = Join-Path $InstallDir $BinaryName
    if (-not (Test-Path $binPath)) {
        Write-Err "Build failed — $binPath not found"
        exit 1
    }

    Write-Ok "Binary installed to $binPath"

    # Check if install dir is in PATH.
    $pathDirs = $env:PATH -split ';'
    if ($InstallDir -notin $pathDirs) {
        Write-Warn "$InstallDir is not in your PATH"
        Write-Warn "Add it: `$env:PATH += `";$InstallDir`""
    }
}

function Invoke-Install {
    if ($env:SKIP_BUILD -eq "1") {
        $binPath = (Get-Command $BinaryName).Source
    } else {
        $binPath = Join-Path $InstallDir $BinaryName
    }

    Write-Info "Running '$BinaryName install'..."
    & $binPath install
}

function Main {
    Write-Host ""
    Write-Host "  orchestratr installer"
    Write-Host "  ====================="
    Write-Host ""

    Test-Prerequisites
    Install-Binary
    Invoke-Install

    Write-Host ""
    Write-Ok "Installation complete!"
    Write-Host ""
    Write-Info "Start the daemon:   orchestratr start"
    Write-Info "Check status:       orchestratr status"
    Write-Info "View config:        orchestratr configure"
    Write-Host ""
}

Main
