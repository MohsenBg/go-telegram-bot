package main

import (
	"tel-bot/internal/db"
	"tel-bot/internal/env"
	"tel-bot/internal/handlers"
	"tel-bot/internal/models"

	"tel-bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	token := env.GetEnvString("BOT_TOKEN", "")
	bot, err := tgbotapi.NewBotAPI(token)
	logger.InitLogger()

	if err != nil {
		panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	db.Connect("telbot.db")

	err = models.CreateUsersTable()
	if err != nil {
		panic(err)
	}

	StartBot(bot, updates)
}

func StartBot(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {

		if update.Message != nil {
			if update.Message.From.IsBot {
				continue
			}
			logger.LogUser(update.Message.Chat.ID, update.Message.From.UserName, update.Message.Text)
			handlers.HandleMessage(bot, update.Message.Chat.ID, update.Message.Text)
		} else if update.CallbackQuery != nil {
			handlers.HandleCallback(bot, update.CallbackQuery)
		}
	}
}
