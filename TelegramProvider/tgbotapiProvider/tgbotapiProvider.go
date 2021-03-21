package tgbotapiProvider

import (
	"context"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"regexp"
)

const ModuleName = "TelegramProviderTgBotApi"

// Implement TelegramProvider interface.
type TelegramModule struct {
	bot *tgbotapi.BotAPI
	Log logger.Logger
	DB  DBProvider.DBProvider
}

// Contain command name and offset.
type Command struct {
	Name   string // Command name.
	Offset uint64 // First command character (after slash) in message text.
}

// Create telegram bot.
// Add created bot and provided logger into provider and return it.
func New(logger logger.Logger, botToken string) (TelegramModule, error) {
	logger = logger.SetModuleName(ModuleName)
	logger.Debug("Initialisation started")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Error(fmt.Sprintf("Initialisation failed - '%v'", err))
		return TelegramModule{}, err
	}

	// Enable bot library debug messages.
	//bot.bot.Debug = true

	logger.Debug(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))
	logger.Debug("Initialisation complete")
	return TelegramModule{
		bot: bot,
		Log: logger,
	}, nil
}

// Set DBProvider.
func (bot TelegramModule) SetDBProvider(db DBProvider.DBProvider) TelegramModule {
	bot.DB = db
	return bot
}

// Listener for Telegram API updates with context control.
func (bot TelegramModule) UpdateListener(ctx context.Context, cancel context.CancelFunc) error {
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
			go messageProcessor(bot, *update.Message)
		case <-ctx.Done():
			log.Printf("Closing signal goroutine")
			return ctx.Err()
		}
	}
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

	// TODO - represent each case as a function
	// Process all provided commands.
	for _, command := range commandList {
		bot.Log.Debug(fmt.Sprintf("Received command '%v' in message '%v'", command.Name, message.Text))
		switch command.Name {
		case "help":
			sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, helpCommandResponse, bot.Log)
		case "firstName":
			log.Printf("'%v' command received. Change FirstName", command)
			err := updateFirstName(bot.DB, command, message.Text, message.Chat.ID)
			if err != nil {
				bot.Log.Error(fmt.Sprintf("Can't update FirstName - '%v'", err))
				sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, invalidFirstNameResponse, bot.Log)
			}
		case "lastName":
			log.Printf("'%v' command received. Change LastName", command)
			err := updateLastName(bot.DB, command, message.Text, message.Chat.ID)
			if err != nil {
				bot.Log.Error(fmt.Sprintf("Can't update FirstName - '%v'", err))
				sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, invalidLastNameResponse, bot.Log)
			}
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

// Extract all commands from received message.
// If entities not received or contain only non-commands, return empty slice.
func extractCommandList(message tgbotapi.Message) []Command {
	// If entities not received return empty slice.
	if message.Entities == nil {
		return make([]Command, 0, 0)
	}

	// Search for commandList.
	commandList := make([]Command, 0, 8)
	entities := *message.Entities
	for _, entity := range entities {
		if entity.Type == "bot_command" {
			firstCharacter := entity.Offset + 1 // Avoid initial slash.
			lastCharacter := entity.Offset + entity.Length
			command := Command{
				Name:   message.Text[firstCharacter:lastCharacter],
				Offset: uint64(firstCharacter),
			}
			commandList = append(commandList, command)
		}
	}

	return commandList
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

// Get russian name from argument.
func getNameFromCommandArgument(text string, commandOffset, commandLength uint64) (string, error) {
	argumentOffset := commandOffset + commandLength + 1
	if int64(len(text))-1-int64(argumentOffset) <= 0 {
		return "", errors.ErrArgumentNotProvided
	}

	reFirstArgumentWithSpaces := regexp.MustCompile(`^\s*\S+`)
	firstArgumentWithSpaces := reFirstArgumentWithSpaces.FindString(text[argumentOffset:])

	reFirstArgument := regexp.MustCompile(`\S+`)
	firstArgument := reFirstArgument.FindString(firstArgumentWithSpaces)

	if len(firstArgumentWithSpaces) < 1 {
		return "", errors.ErrArgumentNotProvided
	}

	reRussianLettersOnly := regexp.MustCompile(`^[абвгдеёжзийклмнопрстуфхцчшщъыьэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ]+$`)
	argument := reRussianLettersOnly.FindString(firstArgument)

	if len(argument) < 1 {
		return "", errors.ErrInvalidArgument
	}

	return argument, nil
}

// Validate argument for /firstName command and write it into DB if valid.
func updateFirstName(db DBProvider.DBProvider, command Command, text string, telegramID int64) error {
	firstName, err := getNameFromCommandArgument(text, command.Offset, uint64(len(command.Name)))
	if err != nil {
		return err
	}

	err = db.BotUserUpdateFirstName(telegramID, firstName)
	if err != nil {
		return err
	}

	return nil
}

// Validate argument for /lastName command and write it into DB if valid.
func updateLastName(db DBProvider.DBProvider, command Command, text string, telegramID int64) error {
	lastName, err := getNameFromCommandArgument(text, command.Offset, uint64(len(command.Name)))
	if err != nil {
		return err
	}

	err = db.BotUserUpdateFirstName(telegramID, lastName)
	if err != nil {
		return err
	}

	return nil
}

// Logic for /subscribe* commands.
func commandSubscribe(bot TelegramModule, message tgbotapi.Message, command Command) {
	DBUserID, err := bot.DB.BotUserGetByTelegramID(message.Chat.ID)
	switch {
	case err == errors.ErrNoUsersFound:
		// If user not found use "start" command behavior.
		commandStart(bot, message)
	case err != nil:
		bot.Log.Error(fmt.Sprintf("while subsscribe user with telegram ID '%v' for '%v' - '%v'",
			message.Chat.ID, command.Name, err))
	}

	// Save subscription and response to user with result.
	err = bot.DB.SubscriptionListAdd(DBUserID, command.Name[9:])
	if err != nil {
		bot.Log.Error(fmt.Sprintf("while subsscribe user with telegram ID '%v' for '%v' - '%v'",
			message.Chat.ID, command.Name, err))
		sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, errorWileSubscribe, bot.Log)
	} else {
		sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, successfulSubscribe, bot.Log)
	}
}

// Logic for /unsubscribe* commands.
func commandUnsubscribe(bot TelegramModule, message tgbotapi.Message, command Command) {
	DBUserID, err := bot.DB.BotUserGetByTelegramID(message.Chat.ID)
	switch {
	case err == errors.ErrNoUsersFound:
		// If user not found use "start" command behavior.
		commandStart(bot, message)
	case err != nil:
		bot.Log.Error(fmt.Sprintf("while unsubsscribe user with telegram ID '%v' for '%v' - '%v'",
			message.Chat.ID, command.Name, err))
	}

	// Remove subscription and response to user with result.
	err = bot.DB.SubscriptionListRemove(DBUserID, command.Name[11:])
	if err != nil {
		bot.Log.Error(fmt.Sprintf("while unsubsscribe user with telegram ID '%v' for '%v' - '%v'",
			message.Chat.ID, command.Name, err))
		sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, errorWileUnsubscribe, bot.Log)
	} else {
		sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, successfulUnsubscribe, bot.Log)
	}
}

// Logic for /start command.
func commandStart(bot TelegramModule, message tgbotapi.Message) {
	sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, startCommandResponse, bot.Log)

	err := bot.DB.BotUserAdd(message.Chat.ID)
	if err != nil && err != errors.ErrNoUsersFound {
		// TODO - add exit program with error
		bot.Log.Error(fmt.Sprintf("Can't create new user - '%v'", err))
	}
}
