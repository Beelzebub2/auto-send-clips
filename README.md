# AutoClipSend

A Golang Wails application that automatically monitors a folder for new video files and sends them to Discord via webhook with custom notifications.

## Features

- 🎬 **Automatic Monitoring**: Watches your configured folder for new video files
- 🔔 **Custom Notifications**: Shows a beautiful notification dialog in the bottom-right corner
- 📤 **Discord Integration**: Sends files to Discord via webhook
- 🎵 **Audio Extraction**: Option to extract and send audio only
- 📏 **Size Management**: Automatically compresses files to stay under configured size limit
- ⚙️ **Configurable**: Easy webhook setup through the UI
- 🖥️ **System Tray**: Runs in the background with minimal UI
- 💾 **Persistent Settings**: Settings and statistics are saved between sessions
- 📊 **Statistics**: Tracks clips sent and other usage metrics

## Prerequisites

1. **Go 1.21+** - [Download Go](https://golang.org/dl/)
2. **Wails v2** - Will be installed automatically by the build script
3. **FFmpeg** - Required for video/audio processing
   - Download from [https://ffmpeg.org/download.html](https://ffmpeg.org/download.html)
   - Make sure `ffmpeg.exe` is in your system PATH

## Installation

### Option 1: Download Pre-built Release (Recommended)
1. Go to the [Releases](https://github.com/Beelzebub2/auto-send-clips/releases) page
2. Download the latest `autoclipsend.exe`
3. Place it in a folder of your choice
4. Run the executable directly

### Option 2: Build from Source
1. **Clone or download** this project to your local machine
2. **Open Command Prompt** and navigate to the project directory
3. **Run the build script**:
   ```cmd
   build.bat
   ```
4. The executable will be created in `build\bin\autoclipsend.exe`

## Releases and Updates

### 📦 Download Pre-built Executable

You can download the latest pre-built executable from the [Releases](https://github.com/Beelzebub2/auto-send-clips/releases) page instead of building from source.

### 🔄 Update Checking

The application automatically checks for updates by comparing the current version with the latest release on GitHub:

1. Click the download link to get the latest version
2. Close the application before installing the update

## Setup

1. **Get Discord Webhook URL**:
   - Go to your Discord server
   - Right-click on the channel where you want to send clips
   - Select "Edit Channel" → "Integrations" → "Webhooks"
   - Create a new webhook and copy the URL

2. **Configure the Application**:
   - Run `autoclipsend.exe`
   - Paste your Discord webhook URL
   - Set your preferred monitor folder path
   - Configure other settings as needed
   - Click "Save Configuration"

3. **Create the Monitor Folder** (if it doesn't exist)

## Usage

1. **Start the Application**: Run `autoclipsend.exe`
2. **Minimize to System Tray**: The app will run in the background
3. **Add Videos**: Place new video files in the monitored folder
4. **Handle Notifications**: When a new video is detected:
   - A notification dialog will appear
   - Enter a custom message (optional)
   - Choose "Audio Only" if you want to extract audio
   - Click "Send to Discord" or "Cancel"

## Supported Video Formats

- `.mp4`
- `.avi`
- `.mov`
- `.mkv`
- `.wmv`
- `.flv`
- `.webm`
- `.m4v`

## File Size Management

- Default maximum file size: **10MB** (configurable)
- If a file exceeds the size limit, it will be automatically compressed
- Audio extraction reduces file size significantly
- Compression maintains reasonable quality while meeting size limits

## Configuration

The application stores its configuration in:
```
%APPDATA%\AutoClipSend\settings.json
```

The settings include:
- Monitor path
- Discord webhook URL
- Maximum file size
- Audio extraction preference
- Notification settings
- Compression settings

## Statistics

The application tracks:
- Total clips sent
- Current session clips
- Total data size
- Uptime and usage metrics

## Troubleshooting

### FFmpeg Not Found
```
Error: ffmpeg not found in PATH
```
**Solution**: Download FFmpeg and add it to your system PATH

### Webhook Errors
```
Error: discord API error: 400
```
**Solution**: Check that your webhook URL is correct and valid

### File Access Errors
```
Error: permission denied
```
**Solution**: Run the application as Administrator if needed

### Build Errors
```
Error: Wails not found
```
**Solution**: The build script should install Wails automatically. If it fails, manually install:
```cmd
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Update Check Errors
```
Error checking for updates
```
**Solution**: Ensure you have internet connectivity and access to GitHub API

## Development

To modify the application:

1. **Frontend**: Edit files in the `frontend` directory
2. **Backend**: Edit Go files (`app.go`, `main.go`, etc.)
3. **Rebuild**: Run `build.bat` or use the Go build command with version flags

## Dependencies

- [Wails v2](https://wails.io/) - Desktop app framework
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system notifications
- FFmpeg - Video/audio processing
- [systray](https://github.com/getlantern/systray) - System tray integration

## License

This project is open source. Feel free to modify and distribute.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Support

If you encounter issues:
1. Check the troubleshooting section
2. Ensure all prerequisites are installed
3. Verify the monitor folder exists
4. Test with a valid Discord webhook URL
5. Check the [GitHub Issues](https://github.com/Beelzebub2/auto-send-clips/issues) page

---

**Created with ❤️ using Wails and Go**
