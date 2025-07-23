# WSL 2 Enablement Script for Windows
# This script enables WSL 2 features on Windows 10/11

param(
    [switch]$Force,
    [switch]$SkipRestart
)

# Function to write colored output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# Function to check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Function to check Windows version
function Test-WindowsVersion {
    $version = [System.Environment]::OSVersion.Version
    $build = $version.Build
    
    if ($build -lt 18362) {
        Write-ColorOutput "❌ Windows 10 version 1903 or higher is required for WSL 2" "Red"
        Write-ColorOutput "Current build: $build" "Yellow"
        return $false
    }
    
    Write-ColorOutput "✅ Windows version is compatible (Build: $build)" "Green"
    return $true
}

# Function to check virtualization support
function Test-VirtualizationSupport {
    Write-ColorOutput "🔍 Checking virtualization support..." "Cyan"
    
    # Check if virtualization is enabled in BIOS
    $virtualization = Get-WmiObject -Class Msvm_VirtualSystemSettingData -Namespace root\virtualization\v2 -ErrorAction SilentlyContinue
    
    if ($virtualization) {
        Write-ColorOutput "✅ Virtualization is enabled" "Green"
        return $true
    } else {
        Write-ColorOutput "⚠️  Virtualization might not be enabled in BIOS" "Yellow"
        Write-ColorOutput "Please enable virtualization (Intel VT-x/AMD-V) in your BIOS settings" "Yellow"
        return $false
    }
}

# Function to enable Windows features
function Enable-WindowsFeatures {
    Write-ColorOutput "🔧 Enabling Windows features for WSL 2..." "Cyan"
    
    $features = @(
        "Microsoft-Windows-Subsystem-Linux",
        "VirtualMachinePlatform"
    )
    
    foreach ($feature in $features) {
        Write-ColorOutput "Enabling $feature..." "Yellow"
        
        try {
            $result = Enable-WindowsOptionalFeature -Online -FeatureName $feature -All -NoRestart
            if ($result.RestartNeeded) {
                Write-ColorOutput "⚠️  Restart required after enabling $feature" "Yellow"
            } else {
                Write-ColorOutput "✅ $feature enabled successfully" "Green"
            }
        }
        catch {
            Write-ColorOutput "❌ Failed to enable $feature : $($_.Exception.Message)" "Red"
            return $false
        }
    }
    
    return $true
}

# Function to download and install WSL 2 kernel update
function Install-WSL2Kernel {
    Write-ColorOutput "📥 Downloading WSL 2 Linux kernel update..." "Cyan"
    
    $kernelUrl = "https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi"
    $kernelPath = "$env:TEMP\wsl_update_x64.msi"
    
    try {
        Invoke-WebRequest -Uri $kernelUrl -OutFile $kernelPath -UseBasicParsing
        Write-ColorOutput "✅ Kernel update downloaded successfully" "Green"
        
        Write-ColorOutput "🔧 Installing WSL 2 kernel update..." "Cyan"
        Start-Process msiexec.exe -Wait -ArgumentList "/I $kernelPath /quiet"
        Write-ColorOutput "✅ WSL 2 kernel update installed successfully" "Green"
        
        # Clean up
        Remove-Item $kernelPath -Force -ErrorAction SilentlyContinue
    }
    catch {
        Write-ColorOutput "❌ Failed to download/install WSL 2 kernel update: $($_.Exception.Message)" "Red"
        return $false
    }
    
    return $true
}

# Function to set WSL 2 as default
function Set-WSL2Default {
    Write-ColorOutput "⚙️  Setting WSL 2 as default version..." "Cyan"
    
    try {
        wsl --set-default-version 2
        Write-ColorOutput "✅ WSL 2 set as default version" "Green"
        return $true
    }
    catch {
        Write-ColorOutput "❌ Failed to set WSL 2 as default: $($_.Exception.Message)" "Red"
        return $false
    }
}

# Function to install a Linux distribution
function Install-LinuxDistribution {
    Write-ColorOutput "🐧 Installing Ubuntu (default Linux distribution)..." "Cyan"
    
    try {
        wsl --install -d Ubuntu
        Write-ColorOutput "✅ Ubuntu installed successfully" "Green"
        Write-ColorOutput "📝 You will need to create a username and password when Ubuntu starts" "Yellow"
        return $true
    }
    catch {
        Write-ColorOutput "❌ Failed to install Ubuntu: $($_.Exception.Message)" "Red"
        Write-ColorOutput "💡 You can manually install a distribution later using: wsl --install -d <distribution>" "Yellow"
        return $false
    }
}

# Function to check WSL status
function Test-WSLStatus {
    Write-ColorOutput "🔍 Checking WSL status..." "Cyan"
    
    try {
        $wslStatus = wsl --status
        Write-ColorOutput "✅ WSL Status:" "Green"
        Write-Host $wslStatus
        
        $wslList = wsl --list --verbose
        Write-ColorOutput "📋 Installed distributions:" "Green"
        Write-Host $wslList
    }
    catch {
        Write-ColorOutput "❌ Failed to check WSL status: $($_.Exception.Message)" "Red"
    }
}

# Main execution
Write-ColorOutput "🚀 WSL 2 Enablement Script for Windows" "Magenta"
Write-ColorOutput "===============================================" "Magenta"

# Check if running as administrator
if (-not (Test-Administrator)) {
    Write-ColorOutput "❌ This script must be run as Administrator" "Red"
    Write-ColorOutput "Please right-click PowerShell and select 'Run as Administrator'" "Yellow"
    exit 1
}

# Check Windows version
if (-not (Test-WindowsVersion)) {
    exit 1
}

# Check virtualization support
if (-not (Test-VirtualizationSupport)) {
    Write-ColorOutput "💡 Please enable virtualization in BIOS and run this script again" "Yellow"
    if (-not $Force) {
        exit 1
    }
}

# Enable Windows features
if (-not (Enable-WindowsFeatures)) {
    Write-ColorOutput "❌ Failed to enable required Windows features" "Red"
    exit 1
}

# Install WSL 2 kernel update
if (-not (Install-WSL2Kernel)) {
    Write-ColorOutput "❌ Failed to install WSL 2 kernel update" "Red"
    exit 1
}

# Set WSL 2 as default
if (-not (Set-WSL2Default)) {
    Write-ColorOutput "❌ Failed to set WSL 2 as default" "Red"
    exit 1
}

# Install Linux distribution
Install-LinuxDistribution

# Check WSL status
Test-WSLStatus

Write-ColorOutput "🎉 WSL 2 setup completed!" "Green"
Write-ColorOutput "===============================================" "Magenta"

if (-not $SkipRestart) {
    Write-ColorOutput "🔄 A system restart is recommended to complete the setup" "Yellow"
    $restart = Read-Host "Do you want to restart now? (y/N)"
    if ($restart -eq 'y' -or $restart -eq 'Y') {
        Restart-Computer -Force
    }
} else {
    Write-ColorOutput "💡 Please restart your system when convenient to complete the setup" "Yellow"
}

Write-ColorOutput "📚 Useful WSL commands:" "Cyan"
Write-ColorOutput "  wsl --list --verbose    - List installed distributions" "White"
Write-ColorOutput "  wsl --install -d <dist> - Install a specific distribution" "White"
Write-ColorOutput "  wsl --update           - Update WSL" "White"
Write-ColorOutput "  wsl --shutdown         - Shutdown all WSL instances" "White" 