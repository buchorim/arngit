@echo off
setlocal enabledelayedexpansion

:: arngit uninstaller
:: Removes arngit from system

echo.
echo  ==============================
echo   arngit Uninstaller
echo  ==============================
echo.

set "INSTALL_DIR=%LOCALAPPDATA%\arngit"
set "BIN_DIR=%INSTALL_DIR%\bin"
set "CONFIG_DIR=%APPDATA%\arngit"

:: Confirm uninstall
set /p CONFIRM="Are you sure you want to uninstall arngit? [y/N]: "
if /i not "%CONFIRM%"=="y" (
    echo Uninstall cancelled.
    exit /b 0
)

:: Remove from PATH
echo [1/3] Removing from PATH...
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "CURRENT_PATH=%%b"

set "NEW_PATH=!CURRENT_PATH:%BIN_DIR%;=!"
set "NEW_PATH=!NEW_PATH:;%BIN_DIR%=!"
set "NEW_PATH=!NEW_PATH:%BIN_DIR%=!"

if not "!NEW_PATH!"=="!CURRENT_PATH!" (
    setx PATH "!NEW_PATH!" >nul
    echo Removed from PATH.
) else (
    echo Not found in PATH.
)

:: Remove binary
echo [2/3] Removing binary...
if exist "%BIN_DIR%\arngit.exe" (
    del /f /q "%BIN_DIR%\arngit.exe"
    echo Removed arngit.exe
)
if exist "%INSTALL_DIR%" (
    rmdir /s /q "%INSTALL_DIR%" 2>nul
)

:: Ask about config
set /p REMOVE_CONFIG="Remove configuration files? [y/N]: "
if /i "%REMOVE_CONFIG%"=="y" (
    echo [3/3] Removing configuration...
    if exist "%CONFIG_DIR%" (
        rmdir /s /q "%CONFIG_DIR%"
        echo Removed config directory.
    )
) else (
    echo [3/3] Keeping configuration files.
)

echo.
echo  ==============================
echo   Uninstall Complete!
echo  ==============================
echo.
echo arngit has been removed from your system.
echo.

pause
