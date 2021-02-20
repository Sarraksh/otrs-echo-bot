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
func (bot TelegramModule) SetDBProvider(db *DBProvider.DBProvider) TelegramModule {
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
			messageProcessor(*update.Message)
		case <-ctx.Done():
			log.Printf("Closing signal goroutine")
			return ctx.Err()
		}
	}

	//if update.Message == nil { // Ignore any non-Message Updates.
	//	continue
	//}

	//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	//msg.ReplyToMessageID = update.Message.MessageID

	//q := *update.EditedMessage.Entities
	//log.Printf("%T", q)

	//bot.Send(msg)

}

//
func messageProcessor(message tgbotapi.Message) {
	commandList := extractCommandList(message)

	// If no commands received stop message processing.
	if len(commandList) == 0 {
		// TODO - add response to user with help message
		return
	}

	// Process all provided commands.
	for _, command := range commandList {
		switch command.Name {
		case "firstName":
			// TODO - add response to user
			// TODO - save data into persistent storage
			log.Printf("'%v' command recived. Set firstName to '%v'", command, message.Text[len("/firstName "):])
		case "lastName":
			// TODO - add response to user
			// TODO - save data into persistent storage
			log.Printf("'%v' command recived. Set lastName to '%v'", command, message.Text[len("/lastName "):])
		case "start":
			// TODO - add response to user with initial help message
			log.Printf("'%v' command recived", command)
		default:
			// TODO - add response to user with help message
			log.Printf("received invalid command - '%v'", command)
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
