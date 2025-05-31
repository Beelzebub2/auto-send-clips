@echo off
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

REM Create directory structure if missing
if not exist "frontend\dist" (
    echo Creating frontend directory structure...
    mkdir "frontend\dist" 2>nul
)
if not exist "frontend\notification" (
    mkdir "frontend\notification" 2>nul
)

REM Build the application
echo Building application
echo This may take a few minutes...
wails build -clean
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
