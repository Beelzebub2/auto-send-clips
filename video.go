package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"autoclipsend/logger"
)

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
	if runtime.GOOS == "windows" {
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

// compressFile compresses the file to fit within size limits
func (a *App) compressFile(inputPath string, isAudio bool) (string, error) {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "_compressed" + filepath.Ext(inputPath)
	var cmd *exec.Cmd
	if isAudio {
		cmd = exec.Command("ffmpeg", "-i", inputPath, "-acodec", "mp3", "-ab", "64k", "-ar", "22050", "-y", outputPath)
	} else {
		cmd = exec.Command("ffmpeg", "-i", inputPath, "-vcodec", "libx264", "-crf", "28", "-preset", "fast", "-vf", "scale=iw/2:ih/2", "-acodec", "aac", "-ab", "64k", "-y", outputPath)
	}
	if err := runFFmpegCommand(cmd); err != nil {
		logger.Error("ffmpeg compression error: %v", err)
		return "", errors.New("ffmpeg compression error")
	}
	return outputPath, nil
}
