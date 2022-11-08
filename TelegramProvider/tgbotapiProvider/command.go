package tgbotapiProvider

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/myErrors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"regexp"
)

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

// Get russian name from argument.
func getNameFromCommandArgument(text string, commandOffset, commandLength uint64) (string, error) {
	argumentOffset := commandOffset + commandLength + 1
	if int64(len(text))-1-int64(argumentOffset) <= 0 {
		return "", myErrors.ErrArgumentNotProvided
	}

	reFirstArgumentWithSpaces := regexp.MustCompile(`^\s*\S+`)
	firstArgumentWithSpaces := reFirstArgumentWithSpaces.FindString(text[argumentOffset:])

	reFirstArgument := regexp.MustCompile(`\S+`)
	firstArgument := reFirstArgument.FindString(firstArgumentWithSpaces)

	if len(firstArgumentWithSpaces) < 1 {
		return "", myErrors.ErrArgumentNotProvided
	}

	reRussianLettersOnly := regexp.MustCompile(`^[абвгдеёжзийклмнопрстуфхцчшщъыьэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ]+$`)
	argument := reRussianLettersOnly.FindString(firstArgument)

	if len(argument) < 1 {
		return "", myErrors.ErrInvalidArgument
	}

	return argument, nil
}

// Logic for /firstName command.
func commandFirstName(bot TelegramModule, message tgbotapi.Message, command Command) {
	bot.Log.Debug(fmt.Sprintf("'%v' command received. Change FirstName", command))
	db := *bot.DB
	err := updateFirstName(db, command, message.Text, message.Chat.ID)
	if err != nil {
		bot.Log.Error(fmt.Sprintf("Can't update FirstName - '%v'", err))
		sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, invalidFirstNameResponse, bot.Log)
	}
}

// Logic for /lastName command.
func commandLastName(bot TelegramModule, message tgbotapi.Message, command Command) {
	bot.Log.Debug(fmt.Sprintf("'%v' command received. Change LastName", command))
	db := *bot.DB
	err := updateLastName(db, command, message.Text, message.Chat.ID)
	if err != nil {
		bot.Log.Error(fmt.Sprintf("Can't update FirstName - '%v'", err))
		sendPlainTextMessageLogErr(bot.bot, message.Chat.ID, invalidLastNameResponse, bot.Log)
	}
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
	db := *bot.DB
	DBUserID, err := db.BotUserGetByTelegramID(message.Chat.ID)
	switch {
	case err == myErrors.ErrNoUsersFound:
		// If user not found use "start" command behavior.
		commandStart(bot, message)
	case err != nil:
		bot.Log.Error(fmt.Sprintf("while subsscribe user with telegram ID '%v' for '%v' - '%v'",
			message.Chat.ID, command.Name, err))
	}

	// Save subscription and response to user with result.
	err = db.SubscriptionListAdd(DBUserID, command.Name[9:])
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
	db := *bot.DB
	DBUserID, err := db.BotUserGetByTelegramID(message.Chat.ID)
	switch {
	case err == myErrors.ErrNoUsersFound:
		// If user not found use "start" command behavior.
		commandStart(bot, message)
	case err != nil:
		bot.Log.Error(fmt.Sprintf("while unsubsscribe user with telegram ID '%v' for '%v' - '%v'",
			message.Chat.ID, command.Name, err))
	}

	// Remove subscription and response to user with result.
	err = db.SubscriptionListRemove(DBUserID, command.Name[11:])
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

	db := *bot.DB
	err := db.BotUserAdd(message.Chat.ID)
	if err != nil && err != myErrors.ErrNoUsersFound {
		// TODO - add exit program with error
		bot.Log.Error(fmt.Sprintf("Can't create new user - '%v'", err))
	}
}
