# AINative Code Installation Script for Windows
# This script installs the latest version of AINative Code

#Requires -Version 5.1

[CmdletBinding()]
param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\AINativeCode",
    [switch]$AddToPath = $true,
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "AINative-Studio/ainative-code"
$BinaryName = "ainative-code.exe"

# Functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Type = "Info"
    )

    $color = switch ($Type) {
        "Success" { "Green" }
        "Warning" { "Yellow" }
        "Error"   { "Red" }
        default   { "Cyan" }
    }

    Write-Host "[$Type] $Message" -ForegroundColor $color
}

function Get-LatestVersion {
    Write-ColorOutput "Fetching latest version..." -Type "Info"

    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"

    try {
        $response = Invoke-RestMethod -Uri $apiUrl -Method Get
        $version = $response.tag_name

        if ([string]::IsNullOrEmpty($version)) {
            throw "Failed to fetch latest version"
        }

        return $version
    }
    catch {
        Write-ColorOutput "Failed to fetch latest version: $_" -Type "Error"
        exit 1
    }
}

function Get-Platform {
    $arch = if ([Environment]::Is64BitOperatingSystem) {
        if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq "Arm64") {
            "arm64"
        } else {
            "x86_64"
        }
    } else {
        Write-ColorOutput "32-bit Windows is not supported" -Type "Error"
        exit 1
    }

    return "Windows_$arch"
}

function Download-File {
    param(
        [string]$Url,
        [string]$OutputPath
    )

    Write-ColorOutput "Downloading from $Url..." -Type "Info"

    try {
        # Use TLS 1.2
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $Url -OutFile $OutputPath -UseBasicParsing
        $ProgressPreference = 'Continue'

        if (-not (Test-Path $OutputPath)) {
            throw "Download failed: file not found at $OutputPath"
        }
    }
    catch {
        Write-ColorOutput "Download failed: $_" -Type "Error"
        exit 1
    }
}

function Get-FileChecksum {
    param(
        [string]$FilePath
    )

    $hash = Get-FileHash -Path $FilePath -Algorithm SHA256
    return $hash.Hash.ToLower()
}

function Verify-Checksum {
    param(
        [string]$FilePath,
        [string]$ExpectedChecksum
    )

    Write-ColorOutput "Verifying checksum..." -Type "Info"

    $actualChecksum = Get-FileChecksum -FilePath $FilePath

    if ($actualChecksum -ne $ExpectedChecksum.ToLower()) {
        Write-ColorOutput "Checksum verification failed!" -Type "Error"
        Write-ColorOutput "Expected: $ExpectedChecksum" -Type "Error"
        Write-ColorOutput "Got:      $actualChecksum" -Type "Error"
        exit 1
    }

    Write-ColorOutput "Checksum verified" -Type "Success"
}

function Extract-Archive {
    param(
        [string]$ArchivePath,
        [string]$DestinationPath
    )

    Write-ColorOutput "Extracting archive..." -Type "Info"

    try {
        if (-not (Test-Path $DestinationPath)) {
            New-Item -ItemType Directory -Path $DestinationPath -Force | Out-Null
        }

        Expand-Archive -Path $ArchivePath -DestinationPath $DestinationPath -Force
    }
    catch {
        Write-ColorOutput "Failed to extract archive: $_" -Type "Error"
        exit 1
    }
}

function Add-ToPath {
    param(
        [string]$Directory
    )

    Write-ColorOutput "Adding to PATH..." -Type "Info"

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($userPath -notlike "*$Directory*") {
        $newPath = "$userPath;$Directory"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

        # Update current session
        $env:Path = "$env:Path;$Directory"

        Write-ColorOutput "Added $Directory to PATH" -Type "Success"
        Write-ColorOutput "You may need to restart your terminal for PATH changes to take effect" -Type "Warning"
    } else {
        Write-ColorOutput "$Directory is already in PATH" -Type "Info"
    }
}

function Test-Installation {
    Write-ColorOutput "Verifying installation..." -Type "Info"

    try {
        $version = & "$InstallDir\$BinaryName" version --short 2>&1

        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "AINative Code $version installed successfully!" -Type "Success"
            return $true
        } else {
            Write-ColorOutput "Installation verification failed" -Type "Warning"
            return $false
        }
    }
    catch {
        Write-ColorOutput "Could not verify installation: $_" -Type "Warning"
        return $false
    }
}

function Main {
    Write-Host ""
    Write-Host "╔════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║      AINative Code Installation Script    ║" -ForegroundColor Cyan
    Write-Host "╚════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""

    # Get platform
    $platform = Get-Platform
    Write-ColorOutput "Detected platform: $platform" -Type "Info"

    # Get version
    if ($Version -eq "latest") {
        $Version = Get-LatestVersion
    }
    Write-ColorOutput "Installing version: $Version" -Type "Info"

    # Construct download URLs
    $versionWithoutV = $Version.TrimStart("v")
    $archiveName = "ainative-code_${versionWithoutV}_${platform}.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$archiveName"
    $checksumsUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "ainative-code-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        # Download archive
        $archivePath = Join-Path $tempDir $archiveName
        Download-File -Url $downloadUrl -OutputPath $archivePath

        # Download checksums
        $checksumsPath = Join-Path $tempDir "checksums.txt"
        Download-File -Url $checksumsUrl -OutputPath $checksumsPath

        # Verify checksum
        $checksumContent = Get-Content $checksumsPath
        $expectedChecksum = ($checksumContent | Select-String $archiveName).ToString().Split()[0]

        if ([string]::IsNullOrEmpty($expectedChecksum)) {
            Write-ColorOutput "Could not find checksum for $archiveName. Skipping verification." -Type "Warning"
        } else {
            Verify-Checksum -FilePath $archivePath -ExpectedChecksum $expectedChecksum
        }

        # Extract archive
        Extract-Archive -ArchivePath $archivePath -DestinationPath $tempDir

        # Create install directory
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        # Copy binary
        $binaryPath = Join-Path $tempDir $BinaryName
        if (-not (Test-Path $binaryPath)) {
            Write-ColorOutput "Binary not found in archive" -Type "Error"
            exit 1
        }

        Write-ColorOutput "Installing to $InstallDir..." -Type "Info"
        Copy-Item -Path $binaryPath -Destination (Join-Path $InstallDir $BinaryName) -Force

        # Add to PATH if requested
        if ($AddToPath) {
            Add-ToPath -Directory $InstallDir
        }

        # Test installation
        Test-Installation

        Write-Host ""
        Write-Host "╔════════════════════════════════════════════╗" -ForegroundColor Green
        Write-Host "║           Installation Complete!           ║" -ForegroundColor Green
        Write-Host "╚════════════════════════════════════════════╝" -ForegroundColor Green
        Write-Host ""
        Write-Host "Get started with:"
        Write-Host "  ainative-code version    # Show version"
        Write-Host "  ainative-code chat       # Start a chat session"
        Write-Host ""
        Write-Host "For more information, visit:"
        Write-Host "  https://github.com/$Repo"
        Write-Host ""
    }
    finally {
        # Cleanup
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Run main function
Main
