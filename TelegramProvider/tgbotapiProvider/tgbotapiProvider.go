package tgbotapiProvider

import (
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

const ModuleName = "TelegramProviderTgBotApi"

type TelegramModule struct {
	bot *tgbotapi.BotAPI
	Log logger.Logger
}

func New(logger logger.Logger, botToken string) (TelegramModule, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return TelegramModule{}, err
	}

	logger.SetModuleName(ModuleName)
	return TelegramModule{
		bot: bot,
		Log: logger,
	}, nil
}

func Update(bot *tgbotapi.BotAPI) {

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("Initialise  updates error - '%v'", err)
		return
	}

	for update := range updates {
		//if update.Message == nil { // Ignore any non-Message Updates.
		//	continue
		//}

		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		//msg.ReplyToMessageID = update.Message.MessageID

		log.Printf("update:\n'%+v'", update)
		if update.Message != nil {
			log.Printf("update.Message:\n'%+v'", update.Message)
			log.Printf("update.Message.Chat:\n'%+v'", update.Message.Chat)
			log.Printf("update.Message.Entities:\n'%+v'", update.Message.Entities)
			messageProcessor(*update.Message)
		} else {
			log.Printf("update.Message:\n'%+v'", nil)
			log.Printf("update.Message.Chat:\n'%+v'", nil)
		}
		if update.EditedMessage != nil {
			log.Printf("update.EditedMessage:\n'%+v'", update.EditedMessage)
			log.Printf("update.EditedMessage.Chat:\n'%+v'", update.EditedMessage.Chat)
			log.Printf("update.EditedMessage.Entities:\n'%+v'", update.EditedMessage.Entities)
			messageProcessor(*update.EditedMessage)

		}

		//q := *update.EditedMessage.Entities
		//log.Printf("%T", q)

		//bot.Send(msg)
	}
}

//
func messageProcessor(message tgbotapi.Message) {
	commandList := extractCommandList(message)

	// If no commands received stop message processing.
	if len(commandList) == 0 {
		// TODO - add response to user with help message
		return
	}

	for _, command := range commandList {
		switch command {
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
func extractCommandList(message tgbotapi.Message) []string {
	// If Entities not received return empty slice
	if message.Entities == nil {
		return make([]string, 0, 0)
	}

	commands := make([]string, 0, 8)
	entities := *message.Entities

	// Search commands
	for _, entity := range entities {
		if entity.Type == "bot_command" {
			firstCharacter := entity.Offset + 1 // Avoid initial slash.
			lastCharacter := entity.Offset + entity.Length
			commands = append(commands, message.Text[firstCharacter:lastCharacter])
		}
	}

	return commands
}
