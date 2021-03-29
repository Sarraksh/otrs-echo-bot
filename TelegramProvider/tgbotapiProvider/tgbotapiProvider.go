package tgbotapiProvider

import (
	"context"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

const ModuleName = "TelegramProviderTgBotApi"

// Implement TelegramProvider interface.
type TelegramModule struct {
	bot *tgbotapi.BotAPI
	Log logger.Logger
	DB  *DBProvider.DBProvider
}

// Contain command name and offset.
type Command struct {
	Name   string // Command name.
	Offset uint64 // First command character (after slash) in message text.
}

// Initialise telegram bot.
// Add created bot and provided logger and DB modules into provider.
func (bot *TelegramModule) Initialise(botToken string, logger logger.Logger, db *DBProvider.DBProvider) error {
	logger = logger.SetModuleName(ModuleName)
	logger.Debug("Initialisation started")
	newBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Error(fmt.Sprintf("Initialisation failed - '%v'", err))
		return err
	}

	// TODO - add option for enable internal bot debug
	// Enable bot library debug messages.
	//bot.bot.Debug = true

	logger.Debug(fmt.Sprintf("Authorized on account %s", newBot.Self.UserName))
	logger.Debug("Initialisation complete")

	bot.bot = newBot
	bot.Log = logger
	bot.DB = db
	return nil
}

// Listener for Telegram API updates with context control.
func (bot *TelegramModule) UpdateListener(ctx context.Context, cancel context.CancelFunc) error {
	// Initialise API listener.
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.bot.GetUpdatesChan(u)
	if err != nil {
		bot.Log.Error(fmt.Sprintf("Get updates error - '%v'", err))
		cancel()
		return err
	}

	// Wait for updates from Telegram API or sigterm.
	for {
		select {
		case update := <-updates:
			go messageProcessor(*bot, *update.Message)
		case <-ctx.Done():
			log.Printf("Closing signal goroutine")
			return ctx.Err()
		}
	}
}

func (bot *TelegramModule) SendEventMessage(chatID int64, text string) error {
	return sendPlainTextMessage(bot.bot, chatID, text)
}

//
func messageProcessor(bot TelegramModule, message tgbotapi.Message) {
	commandList := extractCommandList(message)

	// If no commands received stop message processing.
	if len(commandList) == 0 {
		bot.Log.Debug(fmt.Sprintf("Received no command in message '%v'", message.Text))
		sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, noCommandInMessage, bot.Log)
		return
	}

	// Process all provided commands.
	for _, command := range commandList {
		bot.Log.Debug(fmt.Sprintf("Received command '%v' in message '%v'", command.Name, message.Text))
		switch command.Name {
		case "help":
			sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, helpCommandResponse, bot.Log)
		case "firstName":
			commandFirstName(bot, message, command)
		case "lastName":
			commandLastName(bot, message, command)
		case "start":
			commandStart(bot, message)
		case "subscribeTeam1", "subscribeTeam2", "subscribeTeam3":
			commandSubscribe(bot, message, command)
		case "unsubscribeTeam1", "unsubscribeTeam2", "unsubscribeTeam3":
			commandUnsubscribe(bot, message, command)
		default:
			sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, invalidCommandResponse, bot.Log)
		}
	}
}

// Send simple text message into provided chat.
func sendPlainTextMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	return err
}

// Send simple text message into provided chat and log error if occurred.
func sendPlainTextMessageLogErr(bot *tgbotapi.BotAPI, chatID int64, text string, Log logger.Logger) {
	err := sendPlainTextMessage(bot, chatID, text)
	if err != nil {
		Log.Error(fmt.Sprintf("Can't send message - '%v'", err))
	}
}
