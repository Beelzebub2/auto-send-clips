@echo off
setlocal enabledelayedexpansion
echo ====================================
echo    Building AutoClipSend
echo ====================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from: https://golang.org/dl/
    pause
    exit /b 1
)

echo ✓ Go is installed
echo.

REM Initialize Go modules
echo Initializing Go modules...
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Failed to initialize Go modules
    pause
    exit /b 1
)
echo ✓ Go modules initialized
echo.

REM Check if Wails is installed
where wails >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Installing Wails CLI...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if %ERRORLEVEL% NEQ 0 (
        echo ERROR: Failed to install Wails
        pause
        exit /b 1
    )
    echo ✓ Wails CLI installed
) else (
    echo ✓ Wails CLI is already installed
)
echo.

REM Check for FFmpeg
where ffmpeg >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: FFmpeg not found in PATH
    echo The application requires FFmpeg for video processing
    echo Please download it from: https://ffmpeg.org/download.html
    echo.
    set /p continue="Continue anyway? (y/n): "
    if /i not "%continue%"=="y" (
        echo Build cancelled
        pause
        exit /b 1
    )
) else (
    echo ✓ FFmpeg is available
)
echo.

REM Build the application
echo Building application
echo This may take a few minutes...

cd frontend
echo Running npm build...
call npm run build
cd ..

REM Get version information for build
echo Preparing version information...

REM Check if this is a Git repository and if there are any tags
for /f "tokens=*" %%i in ('git describe --tags --abbrev=0 2^>nul') do set GIT_VERSION=%%i

echo.
echo ====================================
echo     Version Configuration
echo ====================================
echo.

if not "%GIT_VERSION%"=="" (
    echo Latest Git tag found: %GIT_VERSION%
    echo.
)

echo Choose version option:
echo 1. Enter custom version (e.g., v1.0.0, v2.1.3)
echo 2. Use Git tag version (if available^)
echo 3. Use development version
echo 4. Use version from VERSION.json
echo.

set /p version_choice="Enter choice (1-4): "

if "%version_choice%"=="1" (
    set /p VERSION="Enter version (e.g., v1.0.0): "
    if "!VERSION!"=="" (
        echo Error: Version cannot be empty
        pause
        exit /b 1
    )
    REM Update VERSION.json with the new version
    echo Updating VERSION.json...
    powershell -Command "$version = '!VERSION!'; $cleanVersion = $version -replace '^v', ''; $versionObj = @{version = $cleanVersion}; $versionObj | ConvertTo-Json | Set-Content -Path 'VERSION.json'"
    echo ✓ VERSION.json updated with version: !VERSION!
) else if "%version_choice%"=="2" (
    if "%GIT_VERSION%"=="" (
        echo Error: No Git tags found. Please create a tag first or choose option 1.
        echo To create a tag: git tag v1.0.0
        pause
        exit /b 1
    )
    set VERSION=%GIT_VERSION%
    REM Update VERSION.json with the git tag version
    echo Updating VERSION.json...
    powershell -Command "$version = '%GIT_VERSION%'; $cleanVersion = $version -replace '^v', ''; $versionObj = @{version = $cleanVersion}; $versionObj | ConvertTo-Json | Set-Content -Path 'VERSION.json'"
    echo ✓ VERSION.json updated with version: %GIT_VERSION%
) else if "%version_choice%"=="3" (
    set VERSION=dev
    echo Using development version - VERSION.json will not be updated
) else if "%version_choice%"=="4" (
    REM Read version from VERSION.json
    for /f "tokens=*" %%i in ('powershell -Command "(Get-Content 'VERSION.json' | ConvertFrom-Json).version"') do set JSON_VERSION=%%i
    if "!JSON_VERSION!"=="" (
        echo Error: Could not read version from VERSION.json
        pause
        exit /b 1
    )
    set VERSION=v!JSON_VERSION!
    echo Using version from VERSION.json: !VERSION!
) else (
    echo Invalid choice. Using development version.
    set VERSION=dev
)

for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set COMMIT=%%i
if "%COMMIT%"=="" set COMMIT=unknown

for /f "tokens=*" %%i in ('powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'"') do set BUILD_DATE=%%i

echo Building with version: %VERSION%
echo Build commit: %COMMIT%
echo Build date: %BUILD_DATE%

REM Build with version information embedded
wails build -clean -ldflags "-X autoclipsend/version.Version=%VERSION% -X autoclipsend/version.Commit=%COMMIT% -X autoclipsend/version.Date=%BUILD_DATE%"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ====================================
    echo        BUILD SUCCESSFUL! 🎉
    echo ====================================
    echo.
    echo Executable location: build\bin\autoclipsend.exe
    echo.
    echo Setup Instructions:
    echo 1. Ensure FFmpeg is installed and in your PATH
    echo 2. Create the folder: E:\Highlights\Clips\Screen Recording
    echo 3. Get your Discord webhook URL
    echo 4. Run the application and configure the webhook
    echo.
    echo The application will run in the background and monitor
    echo the folder for new video files automatically.
    echo Version %VERSION% has been embedded in the executable.
    echo.
) else (
    echo.
    echo ====================================
    echo         BUILD FAILED! ❌
    echo ====================================
    echo.
    echo Common solutions:
    echo 1. Ensure Go is properly installed
    echo 2. Check your internet connection
    echo 3. Make sure you have write permissions
    echo 4. Try running as Administrator
    echo.
)

pause
