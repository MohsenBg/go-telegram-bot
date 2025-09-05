package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"tel-bot/internal/env"
	"tel-bot/internal/logger"
	"tel-bot/internal/models"
	"tel-bot/internal/validation"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xuri/excelize/v2"
)

// ------------------- Step Messages -------------------
var StepMessages = map[int]string{
	0: "🎉 سلام! به ربات ثبت نام جشن فارغ‌التحصیلی خوش آمدید.",
	1: "✏️ لطفاً نام و نام خانوادگی خود را وارد کنید.",
	2: "📞 لطفاً شماره تلفن خود را وارد کنید.",
	3: "👥 تعداد نفرات همراه خود را وارد کنید.",
	4: "🎓 رشته تحصیلی خود را وارد کنید.",
	5: "🆔 شماره دانشجویی خود را وارد کنید.",
	6: "💳 شماره تراکنش پرداخت خود را وارد کنید.",
	7: "✅ لطفاً اطلاعات خود را تأیید کنید.",
	8: "🎊 ثبت نام شما با موفقیت انجام شد. ممنون!",
	9: "👏 کاربر گرامی، اطلاعات شما با موفقیت ثبت شد. در صورت نیاز به ثبت اطلاعات جدید، لطفاً بات را با دستور /start دوباره راه‌اندازی کنید.",
}

// ------------------- In-Memory Chat Storage -------------------
var usersChat = make(map[int64]*models.UserChat)

// ------------------- Password Protection -------------------
const defaultPassword = "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"

var waitingForPassword = make(map[int64]bool)

// ------------------- Public Handler -------------------
func HandleMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	text = strings.TrimSpace(text)

	// Restart on /start
	if text == "/start" {
		delete(usersChat, chatID)
		usersChat[chatID] = &models.UserChat{Step: 1}
		sendAndLog(bot, chatID, StepMessages[0])
		sendAndLog(bot, chatID, StepMessages[1])
		return
	}

	// Protected Excel export
	if text == "/get_users" {
		waitingForPassword[chatID] = true
		sendAndLog(bot, chatID, "🔑 لطفاً رمز عبور را وارد کنید:")
		return
	}

	// Check if user is entering password
	if waitingForPassword[chatID] {
		if hashPassword(text) == env.GetEnvString("PASSWORD", defaultPassword) {
			delete(waitingForPassword, chatID)
			SendUsersExcel(bot, chatID)
		} else {
			sendAndLog(bot, chatID, "❌ رمز عبور اشتباه است")
		}
		return
	}

	if _, exists := usersChat[chatID]; !exists {
		usersChat[chatID] = &models.UserChat{Step: 1}
		sendAndLog(bot, chatID, StepMessages[0])
		sendAndLog(bot, chatID, StepMessages[1])
		return
	}

	userChat := usersChat[chatID]

	switch userChat.Step {
	case 1:
		handleStep(bot, chatID, text, userChat, handleName)
	case 2:
		handleStep(bot, chatID, text, userChat, handlePhone)
	case 3:
		promptCompanions(bot, chatID)
	case 4:
		promptMajor(bot, chatID)
	case 5:
		handleStep(bot, chatID, text, userChat, handleStudentID)
	case 6:
		handleStep(bot, chatID, text, userChat, handleTransaction)
	case 9:
		sendAndLog(bot, chatID, StepMessages[9])
	default:
		sendAndLog(bot, chatID, "⚠️ خطا: لطفا ربات دوباره راه اندازی کنید با /start")
	}
}

// ------------------- Step Handler Helper -------------------
type stepHandlerFunc func(*tgbotapi.BotAPI, int64, string, *models.UserChat) error

func handleStep(bot *tgbotapi.BotAPI, chatID int64, text string, userChat *models.UserChat, handler stepHandlerFunc) {
	if err := handler(bot, chatID, text, userChat); err != nil {
		sendAndLog(bot, chatID, "❌ "+err.Error())
		sendAndLog(bot, chatID, StepMessages[userChat.Step])
	}
}

// ------------------- Step Handlers -------------------
func handleName(bot *tgbotapi.BotAPI, chatID int64, text string, userChat *models.UserChat) error {
	if err := validation.ValidateName(text); err != nil {
		return err
	}
	userChat.Information.Name = text
	userChat.Step = 2
	sendAndLog(bot, chatID, "✅ نام شما با موفقیت ثبت شد.")
	sendAndLog(bot, chatID, StepMessages[2])
	return nil
}

func handlePhone(bot *tgbotapi.BotAPI, chatID int64, text string, userChat *models.UserChat) error {
	if err := validation.ValidatePhone(text); err != nil {
		return err
	}
	userChat.Information.Phone = text
	userChat.Step = 3
	sendAndLog(bot, chatID, "✅ شماره تلفن با موفقیت ثبت شد.")
	promptCompanions(bot, chatID)
	return nil
}

func handleStudentID(bot *tgbotapi.BotAPI, chatID int64, text string, userChat *models.UserChat) error {
	if err := validation.ValidateStudentID(text); err != nil {
		return err
	}
	userChat.Information.StudentID = text
	userChat.Step = 6
	sendAndLog(bot, chatID, "🆔 شماره دانشجویی ثبت شد.")
	sendAndLog(bot, chatID, StepMessages[6])
	return nil
}

func handleTransaction(bot *tgbotapi.BotAPI, chatID int64, text string, userChat *models.UserChat) error {
	if err := validation.ValidateTransaction(text); err != nil {
		return err
	}
	userChat.Information.Transaction = text
	userChat.Step = 7

	info := userChat.Information
	summary := "📋 لطفاً اطلاعات خود را بررسی کنید:\n" +
		"👤 نام: " + info.Name + "\n" +
		"📞 شماره تلفن: " + info.Phone + "\n" +
		"👥 نفرات همراه: " + strconv.Itoa(info.Companions) + "\n" +
		"🎓 رشته تحصیلی: " + info.Major + "\n" +
		"🆔 شماره دانشجویی: " + info.StudentID + "\n" +
		"💳 شماره تراکنش: " + info.Transaction + "\n\n" +
		"✅ آیا اطلاعات صحیح است؟"

	msg := tgbotapi.NewMessage(chatID, summary)
	msg.ReplyMarkup = confirmationKeyboard()
	bot.Send(msg)
	logger.LogBot(chatID, summary)
	return nil
}

// ------------------- Buttons -------------------
func confirmationKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تایید", "confirm"),
			tgbotapi.NewInlineKeyboardButtonData("❌ انصراف", "cancel"),
		),
	)
}

func companionsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("0️⃣", "companions_0"),
			tgbotapi.NewInlineKeyboardButtonData("1️⃣", "companions_1"),
			tgbotapi.NewInlineKeyboardButtonData("2️⃣", "companions_2"),
			tgbotapi.NewInlineKeyboardButtonData("3️⃣", "companions_3"),
			tgbotapi.NewInlineKeyboardButtonData("4️⃣", "companions_4"),
			tgbotapi.NewInlineKeyboardButtonData("5️⃣", "companions_5"),
		),
	)
}

func majorKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💻 کامپیوتر", "major_کامپیوتر"),
			tgbotapi.NewInlineKeyboardButtonData("⚡ برق", "major_برق"),
		),
	)
}

// ------------------- Prompt Functions -------------------
func promptCompanions(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, StepMessages[3])
	msg.ReplyMarkup = companionsKeyboard()
	bot.Send(msg)
	logger.LogBot(chatID, StepMessages[3])
}

func promptMajor(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, StepMessages[4])
	msg.ReplyMarkup = majorKeyboard()
	bot.Send(msg)
	logger.LogBot(chatID, StepMessages[4])
}

// ------------------- Handle Callback -------------------
func HandleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	userChat, exists := usersChat[chatID]
	if !exists {
		return
	}

	data := callback.Data

	switch {
	case strings.HasPrefix(data, "companions_"):
		count, _ := strconv.Atoi(strings.TrimPrefix(data, "companions_"))
		userChat.Information.Companions = count
		userChat.Step = 4
		sendAndLog(bot, chatID, "👥 تعداد نفرات همراه ثبت شد.")
		promptMajor(bot, chatID)

	case strings.HasPrefix(data, "major_"):
		major := strings.TrimPrefix(data, "major_")
		userChat.Information.Major = major
		userChat.Step = 5
		sendAndLog(bot, chatID, "🎓 رشته تحصیلی ثبت شد.")
		sendAndLog(bot, chatID, StepMessages[5])

	case data == "confirm":
		user := userChat.Information
		user.Username = callback.From.UserName
		user.ChatID = chatID

		if err := user.Save(); err != nil {
			sendAndLog(bot, chatID, "⚠️ خطا در ذخیره اطلاعات. لطفاً دوباره تلاش کنید.")
			return
		}

		sendAndLog(bot, chatID, StepMessages[8])
		userChat.Step = 9

	case data == "cancel":
		delete(usersChat, chatID)
		usersChat[chatID] = &models.UserChat{Step: 1}
		sendAndLog(bot, chatID, "❌ اطلاعات پاک شد. لطفاً دوباره نام و نام خانوادگی خود را وارد کنید.")
		sendAndLog(bot, chatID, StepMessages[1])
	}

	// Acknowledge callback
	bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	bot.Request(tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID))
}

// ------------------- Excel Export -------------------
func SendUsersExcel(bot *tgbotapi.BotAPI, chatID int64) {
	users, err := models.GetAllUsers()
	if err != nil {
		sendAndLog(bot, chatID, "❌ خطا در دریافت اطلاعات کاربران")
		return
	}

	f := excelize.NewFile()
	sheet := "Users"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	rtl := true
	f.SetSheetView(sheet, -1, &excelize.ViewOptions{RightToLeft: &rtl})

	headers := []string{
		"نام و نام خانوادگی",
		"شماره همراه",
		"تعداد همراه",
		"رشته",
		"کد دانشجویی",
		"شماره تراکنش بانکی",
		"نام کاربری ثبت کننده",
		"زمان ثبت",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for rowIndex, u := range users {
		values := []interface{}{
			u.Name,
			u.Phone,
			u.Companions,
			u.Major,
			u.StudentID,
			u.Transaction,
			u.Username,
			u.CreatedAt, // string from DB
		}
		for colIndex, v := range values {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			f.SetCellValue(sheet, cell, v)
			f.SetCellStyle(sheet, cell, cell, cellStyle)
		}
		f.SetRowHeight(sheet, rowIndex+2, 20)
	}

	for i := 1; i <= len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, col, col, 22)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		sendAndLog(bot, chatID, "❌ خطا در ایجاد فایل اکسل")
		return
	}

	fileBytes := tgbotapi.FileBytes{
		Name:  "users_list.xlsx",
		Bytes: buf.Bytes(),
	}

	sendAndLog(bot, chatID, fmt.Sprintf("تعداد رکورد: %d", len(users)))
	bot.Send(tgbotapi.NewDocument(chatID, fileBytes))
	sendAndLog(bot, chatID, "✅ فایل اکسل کاربران ارسال شد.")
}

// ------------------- Password Hash -------------------
func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// ------------------- sendAndLog helper -------------------
func sendAndLog(bot *tgbotapi.BotAPI, chatID int64, text string) {
	bot.Send(tgbotapi.NewMessage(chatID, text))
	logger.LogBot(chatID, text)
}
