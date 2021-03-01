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
func (bot TelegramModule) Update(ctx context.Context, cancel context.CancelFunc) error {
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
		// TODO - add response to user with help message
		return
	}

	// Process all provided commands.
	for _, command := range commandList {
		bot.Log.Debug(fmt.Sprintf("Received command '%v' in message '%v'", command.Name, message.Text))
		switch command.Name {
		case "help":
			err := sendPlainTextMessage(bot.bot, message.Chat.ID, helpCommandResponse)
			if err != nil {
				bot.Log.Error(fmt.Sprintf("Can't sent message - '%v'", err))
			}
		case "firstName":
			log.Printf("'%v' command received. Change FirstName", command)
			err := updateFirstName(bot.DB, command, message.Text, message.Chat.ID)
			if err != nil {
				bot.Log.Error(fmt.Sprintf("Can't update FirstName - '%v'", err))
				err := sendPlainTextMessage(bot.bot, message.Chat.ID, invalidFirstNameResponse)
				if err != nil {
					bot.Log.Error(fmt.Sprintf("Can't sent message - '%v'", err))
				}
			}
		case "lastName":
			log.Printf("'%v' command received. Change LastName", command)
			err := updateFirstName(bot.DB, command, message.Text, message.Chat.ID)
			if err != nil {
				bot.Log.Error(fmt.Sprintf("Can't update FirstName - '%v'", err))
				err := sendPlainTextMessage(bot.bot, message.Chat.ID, invalidLastNameResponse)
				if err != nil {
					bot.Log.Error(fmt.Sprintf("Can't sent message - '%v'", err))
				}
			}
		case "start":
			err := sendPlainTextMessage(bot.bot, message.Chat.ID, startCommandResponse)
			if err != nil {
				bot.Log.Error(fmt.Sprintf("Can't sent message - '%v'", err))
			}
		default:
			err := sendPlainTextMessage(bot.bot, message.Chat.ID, invalidCommandResponse)
			if err != nil {
				bot.Log.Error(fmt.Sprintf("Can't sent message - '%v'", err))
			}
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
