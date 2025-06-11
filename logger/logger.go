package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	fileLogger    *log.Logger
	logFile       *os.File
	once          sync.Once
	logLevel      LogLevel = INFO
	timeFormat             = "2006-01-02 15:04:05"
	enableConsole          = true
)

// Color codes for console output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[37m"
)

// Init initializes the logger with a log file in the user's home directory
func Init() error {
	var err error
	once.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return
		}

		// Create logs directory if it doesn't exist
		logsDir := filepath.Join(homeDir, ".autoclipsend", "logs")
		if err = os.MkdirAll(logsDir, 0755); err != nil {
			return
		}

		// Create or open log file with current date in HOME/.autoclipsend/logs/
		currentDate := time.Now().Format("2006-01-02")
		logFilePath := filepath.Join(logsDir, fmt.Sprintf("autoclipsend_%s.log", currentDate))
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}

		fileLogger = log.New(logFile, "", 0)
	})
	return err
}

// SetLogLevel sets the minimum log level to be logged
func SetLogLevel(level LogLevel) {
	logLevel = level
}

// SetConsoleOutput enables or disables console output
func SetConsoleOutput(enabled bool) {
	enableConsole = enabled
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func getAppDataPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir, nil
}

func writeLog(level LogLevel, format string, v ...interface{}) {
	if level < logLevel {
		return
	}

	// Get caller information
	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)

	// Format the message
	levelStr := []string{"DEBUG", "INFO ", "WARN ", "ERROR"}[level]
	timestamp := time.Now().Format(timeFormat)
	msg := fmt.Sprintf(format, v...)

	// File log format
	fileLogMsg := fmt.Sprintf("[%s] %s [%s:%d] %s", levelStr, timestamp, file, line, msg)

	// Console log format with colors
	var consoleColor string
	switch level {
	case DEBUG:
		consoleColor = colorGray
	case INFO:
		consoleColor = colorBlue
	case WARN:
		consoleColor = colorYellow
	case ERROR:
		consoleColor = colorRed
	}

	consoleLogMsg := fmt.Sprintf("%s[%s]%s %s %s[%s:%d]%s %s",
		consoleColor, levelStr, colorReset,
		timestamp,
		colorGray, file, line, colorReset,
		msg)

	// Log to file
	if fileLogger != nil {
		fileLogger.Println(fileLogMsg)
	}

	// Log to console if enabled
	if enableConsole {
		fmt.Println(consoleLogMsg)
	}
}

func Debug(format string, v ...interface{}) {
	writeLog(DEBUG, format, v...)
}

func Info(format string, v ...interface{}) {
	writeLog(INFO, format, v...)
}

func Warn(format string, v ...interface{}) {
	writeLog(WARN, format, v...)
}

func Error(format string, v ...interface{}) {
	writeLog(ERROR, format, v...)
}
