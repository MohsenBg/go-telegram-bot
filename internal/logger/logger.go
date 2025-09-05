package logger

import (
	"log"
	"os"
	"time"
)

var (
	UserBotLogger *log.Logger
)

func InitLogger() {
	file, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("cannot open log file:", err)
	}

	UserBotLogger = log.New(file, "", 0)
}

// Helper to log user messages with timestamp
func LogUser(chatID int64, username, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	UserBotLogger.Printf("%s [USER] ChatID: %d | Username: %s | Message: %s\n", timestamp, chatID, username, message)
}

// Helper to log bot responses with timestamp
func LogBot(chatID int64, response string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	UserBotLogger.Printf("%s [BOT] ChatID: %d | Response: %s\n", timestamp, chatID, response)
}
