package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"syscall"

	"autoclipsend/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProgressInfo represents the progress of file processing
type ProgressInfo struct {
	Stage       string  `json:"stage"`
	Progress    float64 `json:"progress"`
	Message     string  `json:"message"`
	IsComplete  bool    `json:"isComplete"`
	Error       string  `json:"error,omitempty"`
}

// isVideoFile checks if the file is a video file
func (a *App) isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v"}
	for _, validExt := range videoExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// handleNewVideo processes a newly detected video file
func (a *App) handleNewVideo(filePath string) {
	_, err := os.Stat(filePath)
	if err != nil {
		logger.Error("Error getting file info: %v", err)
		return
	}

	fileName := filepath.Base(filePath)
	logger.Info("Triggering notification for: %s", fileName)
	go a.ShowNotification(fileName, filePath)
}

// Helper to run ffmpeg without showing a console window (Windows only)
func runFFmpegCommand(cmd *exec.Cmd) error {
	if goruntime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	return cmd.Run()
}

// extractAudio extracts audio from video file using ffmpeg
func (a *App) extractAudio(videoPath string) (string, error) {
	outputPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + "_audio.mp3"
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vn", "-acodec", "mp3", "-ab", "128k", "-ar", "44100", "-y", outputPath)
	if err := runFFmpegCommand(cmd); err != nil {
		logger.Error("ffmpeg error: %v", err)
		return "", errors.New("ffmpeg error")
	}
	return outputPath, nil
}

// compressFile compresses the file to fit within size limits using aggressive multi-pass compression
func (a *App) compressFile(inputPath string, isAudio bool) (string, error) {
	maxSizeMB := a.config.MaxFileSize
	maxSizeBytes := maxSizeMB * 1024 * 1024
	
	if isAudio {
		return a.compressAudioAggressively(inputPath, maxSizeBytes)
	}
	
	return a.compressVideoAggressively(inputPath, maxSizeBytes)
}

// compressAudioAggressively compresses audio using multiple passes until target size is reached
func (a *App) compressAudioAggressively(inputPath string, maxSizeBytes int64) (string, error) {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "_compressed.mp3"
	
	// Audio compression settings from highest to lowest quality
	audioSettings := []struct {
		bitrate   string
		sampleRate string
		channels  string
	}{
		{"128k", "44100", "2"},  // Standard quality
		{"96k", "44100", "2"},   // Good quality
		{"64k", "22050", "2"},   // Medium quality
		{"48k", "22050", "2"},   // Lower quality
		{"32k", "22050", "1"},   // Low quality mono
		{"24k", "16000", "1"},   // Very low quality
		{"16k", "11025", "1"},   // Minimum quality
	}
	
	for i, setting := range audioSettings {
		tempPath := outputPath
		if i > 0 {
			tempPath = strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + fmt.Sprintf("_temp_%d.mp3", i)
		}
		
		cmd := exec.Command("ffmpeg", "-i", inputPath, "-acodec", "mp3", "-ab", setting.bitrate, "-ar", setting.sampleRate, "-ac", setting.channels, "-y", tempPath)
		
		if err := runFFmpegCommand(cmd); err != nil {
			logger.Warn("Audio compression attempt %d failed: %v", i+1, err)
			continue
		}
		
		// Check if file size is acceptable
		if fileInfo, err := os.Stat(tempPath); err == nil && fileInfo.Size() <= maxSizeBytes {
			if tempPath != outputPath {
				// Move temp file to final output path
				os.Rename(tempPath, outputPath)
			}
			logger.Info("Audio compressed successfully with setting %d, size: %d bytes", i+1, fileInfo.Size())
			return outputPath, nil
		}
		
		// Clean up temp file if it's not the final output
		if tempPath != outputPath {
			os.Remove(tempPath)
		}
	}
	
	return "", errors.New("could not compress audio to target size")
}

// compressVideoAggressively compresses video using resolution reduction and moderate quality settings
func (a *App) compressVideoAggressively(inputPath string, maxSizeBytes int64) (string, error) {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "_compressed.mp4"
	
	// Get video information first
	videoDuration, err := a.getVideoDuration(inputPath)
	if err != nil {
		logger.Warn("Could not get video duration, using default compression: %v", err)
		return a.fallbackVideoCompression(inputPath, outputPath)
	}
	
	// Calculate target bitrate based on duration and max size
	// Leave some margin for audio and container overhead (20% margin)
	targetBitrate := int64(float64(maxSizeBytes) * 0.8 * 8 / videoDuration) // bits per second
	
	// Resolution-focused compression strategies - prioritize watchable quality
	compressionStrategies := []struct {
		codec       string
		preset      string
		crf         string
		scale       string
		fps         string
		audioBitrate string
		audioRate   string
		customArgs  []string
		description string
	}{
		// Full resolution strategies with good quality
		{"libx264", "fast", "23", "", "fps=30", "128k", "44100", []string{}, "Full resolution, 30fps"},
		{"libx264", "fast", "25", "", "fps=30", "96k", "44100", []string{}, "Full resolution, good quality"},
		
		// 720p strategies (most common sweet spot)
		{"libx264", "fast", "23", "scale=1280:720", "fps=30", "96k", "44100", []string{}, "720p, 30fps"},
		{"libx264", "fast", "25", "scale=1280:720", "fps=30", "64k", "22050", []string{}, "720p, standard quality"},
		
		// 540p strategies (good compromise)
		{"libx264", "fast", "23", "scale=960:540", "fps=30", "64k", "22050", []string{}, "540p, 30fps"},
		{"libx264", "fast", "25", "scale=960:540", "fps=24", "48k", "22050", []string{}, "540p, 24fps"},
		
		// 480p strategies (still very watchable)
		{"libx264", "fast", "23", "scale=854:480", "fps=30", "48k", "22050", []string{}, "480p, 30fps"},
		{"libx264", "fast", "25", "scale=854:480", "fps=24", "48k", "22050", []string{}, "480p, 24fps"},
		
		// 360p strategies (mobile quality)
		{"libx264", "fast", "23", "scale=640:360", "fps=24", "32k", "22050", []string{}, "360p, 24fps"},
		{"libx264", "fast", "25", "scale=640:360", "fps=20", "32k", "22050", []string{}, "360p, 20fps"},
		
		// 240p strategies (last resort but still watchable)
		{"libx264", "veryfast", "25", "scale=426:240", "fps=20", "32k", "22050", []string{}, "240p, 20fps"},
		{"libx264", "veryfast", "27", "scale=426:240", "fps=15", "24k", "16000", []string{}, "240p, 15fps"},
	}
	
	// If target bitrate is very low, use bitrate-based compression instead
	if targetBitrate < 300000 { // less than 300kbps
		return a.compressVideoByBitrate(inputPath, outputPath, targetBitrate, maxSizeBytes)
	}
	
	totalStrategies := len(compressionStrategies)
	for i, strategy := range compressionStrategies {
		// Emit progress update
		progress := float64(i) / float64(totalStrategies)
		a.emitProgress(ProgressInfo{
			Stage:    "compression",
			Progress: progress,
			Message:  fmt.Sprintf("Trying %s compression...", strategy.description),
		})
		
		tempPath := outputPath
		if i > 0 {
			tempPath = strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + fmt.Sprintf("_temp_%d.mp4", i)
		}
		
		// Build ffmpeg command
		args := []string{"-i", inputPath, "-c:v", strategy.codec, "-preset", strategy.preset, "-crf", strategy.crf}
		
		// Add video filters
		var videoFilters []string
		if strategy.scale != "" {
			videoFilters = append(videoFilters, strategy.scale)
		}
		if strategy.fps != "" {
			videoFilters = append(videoFilters, strategy.fps)
		}
		if len(videoFilters) > 0 {
			args = append(args, "-vf", strings.Join(videoFilters, ","))
		}
		
		// Add audio settings
		args = append(args, "-c:a", "aac", "-b:a", strategy.audioBitrate, "-ar", strategy.audioRate)
		
		// Add custom arguments
		args = append(args, strategy.customArgs...)
		
		// Add output path and overwrite flag
		args = append(args, "-y", tempPath)
		
		logger.Info("Attempting compression with: %s", strategy.description)
		cmd := exec.Command("ffmpeg", args...)
		
		if err := runFFmpegCommand(cmd); err != nil {
			logger.Warn("Video compression attempt %d failed: %v", i+1, err)
			// Clean up temp file
			os.Remove(tempPath)
			continue
		}
		
		// Check if file size is acceptable
		if fileInfo, err := os.Stat(tempPath); err == nil && fileInfo.Size() <= maxSizeBytes {
			if tempPath != outputPath {
				// Move temp file to final output path
				os.Rename(tempPath, outputPath)
			}
			
			// Calculate size reduction
			originalInfo, _ := os.Stat(inputPath)
			compressionRatio := float64(fileInfo.Size()) / float64(originalInfo.Size()) * 100
			
			logger.Info("Video compressed successfully with %s, size: %d bytes (%.1f%% of original)", 
				strategy.description, fileInfo.Size(), compressionRatio)
			
			// Emit completion
			a.emitProgress(ProgressInfo{
				Stage:      "compression",
				Progress:   1.0,
				Message:    fmt.Sprintf("Compressed to %s (%.1f%% of original size)", strategy.description, compressionRatio),
				IsComplete: true,
			})
			
			return outputPath, nil
		}
		
		// Clean up temp file if it's not the final output
		if tempPath != outputPath {
			os.Remove(tempPath)
		}
	}
	
	// If all strategies failed, try bitrate-based compression as last resort
	logger.Warn("All CRF-based strategies failed, trying bitrate-based compression")
	return a.compressVideoByBitrate(inputPath, outputPath, targetBitrate/2, maxSizeBytes)
}

// compressVideoByBitrate uses target bitrate for compression
func (a *App) compressVideoByBitrate(inputPath, outputPath string, targetBitrate, maxSizeBytes int64) (string, error) {
	// Multiple bitrate attempts, each one more aggressive
	bitrates := []int64{
		targetBitrate,
		targetBitrate / 2,
		targetBitrate / 3,
		targetBitrate / 4,
		targetBitrate / 6,
		100000, // 100kbps minimum
	}
	
	scales := []string{"", "scale=iw*0.8:ih*0.8", "scale=iw*0.6:ih*0.6", "scale=iw*0.5:ih*0.5", "scale=iw*0.4:ih*0.4", "scale=iw*0.3:ih*0.3"}
	fpsList := []string{"", "fps=30", "fps=24", "fps=20", "fps=15", "fps=10"}
	
	for i, bitrate := range bitrates {
		tempPath := outputPath
		if i > 0 {
			tempPath = strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + fmt.Sprintf("_bitrate_%d.mp4", i)
		}
		
		scale := ""
		fps := ""
		if i < len(scales) {
			scale = scales[i]
		}
		if i < len(fpsList) {
			fps = fpsList[i]
		}
		
		args := []string{"-i", inputPath, "-c:v", "libx264", "-b:v", fmt.Sprintf("%d", bitrate), "-preset", "veryfast", "-maxrate", fmt.Sprintf("%d", bitrate*2), "-bufsize", fmt.Sprintf("%d", bitrate*4)}
		
		// Add video filters
		var videoFilters []string
		if scale != "" {
			videoFilters = append(videoFilters, scale)
		}
		if fps != "" {
			videoFilters = append(videoFilters, fps)
		}
		if len(videoFilters) > 0 {
			args = append(args, "-vf", strings.Join(videoFilters, ","))
		}
		
		// Add audio settings
		audioBitrate := "32k"
		if bitrate > 500000 {
			audioBitrate = "64k"
		}
		args = append(args, "-c:a", "aac", "-b:a", audioBitrate, "-ar", "22050")
		
		// Add output path and overwrite flag
		args = append(args, "-y", tempPath)
		
		cmd := exec.Command("ffmpeg", args...)
		
		if err := runFFmpegCommand(cmd); err != nil {
			logger.Warn("Bitrate compression attempt %d failed: %v", i+1, err)
			os.Remove(tempPath)
			continue
		}
		
		// Check if file size is acceptable
		if fileInfo, err := os.Stat(tempPath); err == nil && fileInfo.Size() <= maxSizeBytes {
			if tempPath != outputPath {
				os.Rename(tempPath, outputPath)
			}
			logger.Info("Video compressed successfully with bitrate %d, size: %d bytes", bitrate, fileInfo.Size())
			return outputPath, nil
		}
		
		if tempPath != outputPath {
			os.Remove(tempPath)
		}
	}
	
	return "", errors.New("could not compress video to target size")
}

// getVideoDuration gets the duration of a video file in seconds
func (a *App) getVideoDuration(inputPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", inputPath)
	if goruntime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	durationStr := strings.TrimSpace(string(output))
	if durationStr == "" {
		return 0, errors.New("could not get duration")
	}
	
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}
	
	return duration, nil
}

// fallbackVideoCompression is a simple fallback compression method
func (a *App) fallbackVideoCompression(inputPath, outputPath string) (string, error) {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-c:v", "libx264", "-crf", "40", "-preset", "veryfast", "-vf", "scale=iw*0.5:ih*0.5,fps=15", "-c:a", "aac", "-b:a", "32k", "-ar", "22050", "-y", outputPath)
	if err := runFFmpegCommand(cmd); err != nil {
		logger.Error("Fallback compression error: %v", err)
		return "", errors.New("fallback compression error")
	}
	return outputPath, nil
}

// emitProgress sends progress updates to the frontend
func (a *App) emitProgress(progress ProgressInfo) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "compressionProgress", progress)
	}
}
