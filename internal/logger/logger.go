package logger

import (
	"log"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	UserBotLogger *log.Logger
)

// InitLogger initializes logger with file rotation
func InitLogger() {
	UserBotLogger = log.New(&lumberjack.Logger{
		Filename:   "bot.log",
		MaxSize:    10, // MB before rotation
		MaxBackups: 3,  // keep last 3 files
		MaxAge:     28, // days
		Compress:   true,
	}, "", 0)
}

// formatTimestamp returns current time as string
func formatTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// LogUser logs messages coming from user
func LogUser(chatID int64, username, message string) {
	UserBotLogger.Printf("%s [USER] ChatID: %d | Username: %s | Message: %s\n",
		formatTimestamp(), chatID, username, message)
}

// LogBot logs messages sent by bot
func LogBot(chatID int64, response string) {
	UserBotLogger.Printf("%s [BOT] ChatID: %d | Response: %s\n",
		formatTimestamp(), chatID, response)
}

