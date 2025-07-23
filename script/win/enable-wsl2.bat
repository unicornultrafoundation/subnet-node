@echo off
setlocal enabledelayedexpansion

REM WSL 2 Enablement Script for Windows (Batch Version)
REM This script enables WSL 2 features on Windows 10/11

echo.
echo ===============================================
echo    WSL 2 Enablement Script for Windows
echo ===============================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] This script must be run as Administrator
    echo Please right-click this file and select "Run as Administrator"
    pause
    exit /b 1
)

echo [INFO] Running as Administrator - OK
echo.

REM Check Windows version
echo [INFO] Checking Windows version...
for /f "tokens=4-5 delims=. " %%i in ('ver') do set VERSION=%%i.%%j
echo [INFO] Windows version: %VERSION%

REM Enable Windows features
echo.
echo [INFO] Enabling Windows features for WSL 2...

echo [INFO] Enabling Windows Subsystem for Linux...
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart
if %errorLevel% neq 0 (
    echo [ERROR] Failed to enable Windows Subsystem for Linux
    pause
    exit /b 1
)

echo [INFO] Enabling Virtual Machine Platform...
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart
if %errorLevel% neq 0 (
    echo [ERROR] Failed to enable Virtual Machine Platform
    pause
    exit /b 1
)

echo [SUCCESS] Windows features enabled successfully
echo.

REM Download and install WSL 2 kernel update
echo [INFO] Downloading WSL 2 Linux kernel update...
powershell -Command "& {Invoke-WebRequest -Uri 'https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi' -OutFile '%TEMP%\wsl_update_x64.msi' -UseBasicParsing}"
if %errorLevel% neq 0 (
    echo [ERROR] Failed to download WSL 2 kernel update
    pause
    exit /b 1
)

echo [INFO] Installing WSL 2 kernel update...
msiexec /i "%TEMP%\wsl_update_x64.msi" /quiet
if %errorLevel% neq 0 (
    echo [ERROR] Failed to install WSL 2 kernel update
    pause
    exit /b 1
)

echo [SUCCESS] WSL 2 kernel update installed successfully
echo.

REM Clean up downloaded file
del "%TEMP%\wsl_update_x64.msi" >nul 2>&1

REM Set WSL 2 as default
echo [INFO] Setting WSL 2 as default version...
wsl --set-default-version 2
if %errorLevel% neq 0 (
    echo [ERROR] Failed to set WSL 2 as default
    pause
    exit /b 1
)

echo [SUCCESS] WSL 2 set as default version
echo.

REM Install Ubuntu distribution
echo [INFO] Installing Ubuntu (default Linux distribution)...
wsl --install -d Ubuntu
if %errorLevel% neq 0 (
    echo [WARNING] Failed to install Ubuntu automatically
    echo [INFO] You can manually install a distribution later using: wsl --install -d ^<distribution^>
) else (
    echo [SUCCESS] Ubuntu installed successfully
    echo [INFO] You will need to create a username and password when Ubuntu starts
)

echo.

REM Check WSL status
echo [INFO] Checking WSL status...
wsl --status
echo.
echo [INFO] Installed distributions:
wsl --list --verbose

echo.
echo ===============================================
echo    WSL 2 setup completed!
echo ===============================================
echo.

REM Ask for restart
set /p restart="Do you want to restart now to complete the setup? (y/N): "
if /i "%restart%"=="y" (
    echo [INFO] Restarting system...
    shutdown /r /t 5 /c "Restarting to complete WSL 2 setup"
) else (
    echo [INFO] Please restart your system when convenient to complete the setup
)

echo.
echo [INFO] Useful WSL commands:
echo   wsl --list --verbose    - List installed distributions
echo   wsl --install -d ^<dist^> - Install a specific distribution
echo   wsl --update           - Update WSL
echo   wsl --shutdown         - Shutdown all WSL instances
echo.

pause 