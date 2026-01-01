@echo off
setlocal enabledelayedexpansion

:: arngit installer
:: Installs arngit to %LOCALAPPDATA%\arngit and adds to PATH

echo.
echo  ==============================
echo   arngit Installer
echo  ==============================
echo.

set "INSTALL_DIR=%LOCALAPPDATA%\arngit"
set "BIN_DIR=%INSTALL_DIR%\bin"
set "EXE_PATH=%BIN_DIR%\arngit.exe"

:: Check if arngit.exe exists in current directory
if not exist "arngit.exe" (
    echo [ERROR] arngit.exe not found in current directory.
    echo Please build the project first:
    echo   go build -o arngit.exe ./cmd/arngit
    exit /b 1
)

:: Create directories
echo [1/4] Creating directories...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

:: Copy executable
echo [2/4] Installing arngit.exe...
copy /Y "arngit.exe" "%EXE_PATH%" >nul
if errorlevel 1 (
    echo [ERROR] Failed to copy arngit.exe
    exit /b 1
)

:: Add to PATH (user level)
echo [3/4] Adding to PATH...
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "CURRENT_PATH=%%b"

echo !CURRENT_PATH! | find /i "%BIN_DIR%" >nul
if errorlevel 1 (
    :: Add to PATH
    setx PATH "!CURRENT_PATH!;%BIN_DIR%" >nul
    if errorlevel 1 (
        echo [WARNING] Failed to add to PATH automatically.
        echo Please add the following directory to your PATH manually:
        echo   %BIN_DIR%
    ) else (
        echo Added to PATH successfully.
    )
) else (
    echo Already in PATH.
)

:: Create config directory
echo [4/4] Creating config directory...
set "CONFIG_DIR=%APPDATA%\arngit"
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"

echo.
echo  ==============================
echo   Installation Complete!
echo  ==============================
echo.
echo Installed to: %EXE_PATH%
echo Config dir:   %CONFIG_DIR%
echo.
echo IMPORTANT: Restart your terminal for PATH changes to take effect.
echo.
echo Quick start:
echo   arngit account add personal
echo   arngit init
echo   arngit push
echo.

pause
