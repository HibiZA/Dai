# Dai CLI Installation Script for Windows
# This PowerShell script downloads and installs the Dai CLI tool for
# dependency management and vulnerability scanning on Windows.

# Ensure we can execute scripts
try {
    $executionPolicy = Get-ExecutionPolicy
    if ($executionPolicy -eq 'Restricted') {
        Write-Host "Changing ExecutionPolicy to RemoteSigned for this process"
        Set-ExecutionPolicy -Scope Process -ExecutionPolicy RemoteSigned
    }
} catch {
    Write-Error "Failed to check or set execution policy: $_"
    exit 1
}

# Default install directory
$installDir = "$env:LOCALAPPDATA\dai"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

# GitHub repo details
$repo = "HibiZA/dai"
$latestReleaseUrl = "https://api.github.com/repos/$repo/releases/latest"

# Get the latest release version
Write-Host "Detecting latest version of Dai CLI..."
try {
    $releaseInfo = Invoke-RestMethod -Uri $latestReleaseUrl
    $latestVersion = $releaseInfo.tag_name
} catch {
    Write-Error "Failed to determine the latest version: $_"
    exit 1
}

if (-not $latestVersion) {
    Write-Error "Could not determine the latest version"
    exit 1
}

Write-Host "Latest version: $latestVersion"

# Construct the download URL
$downloadUrl = "https://github.com/$repo/releases/download/$latestVersion/dai_windows_$arch.zip"
Write-Host "Download URL: $downloadUrl"

# Create a temporary directory
$tempDir = Join-Path $env:TEMP ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

# Download the zip file
Write-Host "Downloading Dai CLI..."
$zipFile = Join-Path $tempDir "dai.zip"
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipFile
} catch {
    Write-Error "Failed to download Dai CLI: $_"
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# Extract the zip file
Write-Host "Extracting..."
try {
    Expand-Archive -Path $zipFile -DestinationPath $tempDir -Force
} catch {
    Write-Error "Failed to extract zip file: $_"
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# Install the binary
Write-Host "Installing to $installDir..."
try {
    Copy-Item -Path (Join-Path $tempDir "dai.exe") -Destination $installDir -Force
} catch {
    Write-Error "Failed to install Dai CLI: $_"
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# Clean up
Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue

# Verify installation
$daiExePath = Join-Path $installDir "dai.exe"
if (Test-Path $daiExePath) {
    Write-Host "Dai CLI installed successfully!"
    
    # Try to get version
    try {
        $version = & $daiExePath version 2>$null
        Write-Host "Version: $version"
    } catch {
        Write-Host "Version: unknown"
    }
    
    Write-Host ""
    Write-Host "To use Dai CLI, run:"
    Write-Host "  dai scan                # Scan for vulnerabilities"
    Write-Host "  dai upgrade [packages]  # Upgrade dependencies"
    Write-Host ""
    Write-Host "For more information, run: dai --help"
} else {
    Write-Error "Installation failed"
    exit 1
}

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath.Contains($installDir)) {
    Write-Host ""
    Write-Host "Adding Dai CLI to your PATH..."
    
    try {
        $newPath = "$userPath;$installDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "Dai CLI added to your PATH successfully!"
        Write-Host "Please restart your terminal for the changes to take effect."
    } catch {
        Write-Host "NOTE: Dai CLI is not in your PATH."
        Write-Host "Please add $installDir to your PATH manually."
        Write-Host "You can also run Dai CLI using its full path: $daiExePath"
    }
} 