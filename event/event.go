package event

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/ClientProvider"
	"github.com/Sarraksh/otrs-echo-bot/DBProvider"
	"github.com/Sarraksh/otrs-echo-bot/Formatter"
	"github.com/Sarraksh/otrs-echo-bot/OTRSProvider"
	"github.com/Sarraksh/otrs-echo-bot/TelegramProvider"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"github.com/Sarraksh/otrs-echo-bot/common/myErrors"
	"sync"
)

// Processing events at all stages.
type Processor struct {
	DB       *DBProvider.DBProvider
	OTRS     *OTRSProvider.OTRSProvider
	Client   *ClientProvider.ClientProvider
	Telegram *TelegramProvider.TelegramProvider
	Log      logger.Logger
	mx       sync.Mutex
}

func (p *Processor) ProcessEvent() {
	// Measure that one event not processing by two routines simultaneously.
	p.mx.Lock()
	defer p.mx.Unlock()

	// Check events for processing.
	p.Log.Debug("Start search for active events.")
	eventDBID, ticketID, err := (*p.DB).OTRSEventGetActive()
	if err == myErrors.ErrNoActiveEvents {
		p.Log.Debug("No active events.")
		return
	}
	if err != nil {
		p.Log.Error(fmt.Sprintf("Can't search active events - '%v'", err))
		return
	}

	// Schedule another processing right after current.
	go p.ProcessEvent() // TODO prevent infinite loop if error occurred

	// Get additional info for event.
	// Current status and detailed info for ticket from OTRS.
	p.Log.Error(fmt.Sprintf("Get active event with eventDBID '%v' and tiketID '%v'", eventDBID, ticketID))
	OTRS := *p.OTRS
	ticketDetails, err := OTRS.GetTicketDetails(ticketID)
	if err != nil {
		// TODO - add logic for close program
		p.Log.Error(fmt.Sprintf("Can't search for active events - '%v'", err))
		return
	}
	status, err := (*p.DB).OTRSEventGetStatus(eventDBID)
	p.Log.Debug(fmt.Sprintf("Processing event with eventDBID '%v' and status '%v'", eventDBID, status))

	// Generate message for bot user.
	var message string
	switch status {
	case "New":
		message = Formatter.EventNewPlainText(ticketDetails)
		p.Log.Debug(fmt.Sprintf("For event with eventDBID '%v' and status '%v' genereted message:\n'%v'", eventDBID, status, message))
	case "Processing", "Suspended":
		message = Formatter.EventReminderPlainText(ticketDetails, p.Log)
		p.Log.Debug(fmt.Sprintf("For event with eventDBID '%v' and status '%v' genereted message:\n'%v'", eventDBID, status, message))
		switch {
		case ticketDetails.Lock == "lock":
			finishEventProcessing("lock", eventDBID, p.DB, p.Log)
		case ticketDetails.StateType == "closed":
			finishEventProcessing("closed", eventDBID, p.DB, p.Log)
		case ticketDetails.StateType == "merged":
			finishEventProcessing("merged", eventDBID, p.DB, p.Log)
			return
		}
	default:
		p.Log.Warning(fmt.Sprintf("For event with eventDBID '%v' recieved unknown status '%v'", eventDBID, status))
		message = Formatter.EventReminderPlainText(ticketDetails, p.Log)
		p.Log.Debug(fmt.Sprintf("For event with eventDBID '%v' and status '%v' genereted message:\n'%v'", eventDBID, status, message))
	}

	// Set status "Processing" for current event.
	p.Log.Debug(fmt.Sprintf("Set status 'Processing' for event with eventDBID '%v'", eventDBID))
	err = (*p.DB).OTRSEventProcessing(eventDBID)
	if err != nil {
		// TODO - add logic for close program
		p.Log.Error(fmt.Sprintf("Can't set 'Processing' for event with eventDBID '%v'", eventDBID))
		return
	}

	// Get team for event by client.
	p.Log.Debug(fmt.Sprintf("Get bounded team for client '%v'", ticketDetails.CustomerID))
	clientModule := *p.Client
	team, err := clientModule.GetTeamByClient(ticketDetails.CustomerID)
	switch err {
	case nil:
		p.Log.Debug(fmt.Sprintf("Team '%v' bounded with client '%v'. Send message to all subscribed users.", team, ticketDetails.CustomerID))
		go sendMessageForByTeam(team, message, p.DB, p.Log, p.Telegram)
	case myErrors.ErrNoTeamBounded:
		p.Log.Debug(fmt.Sprintf("No team bounded with client '%v'. Send message to all users.", ticketDetails.CustomerID))
		go sendMessageForAllBotUsers(message, p.DB, p.Log, p.Telegram)
	case myErrors.ErrClientNotExists:
		p.Log.Debug(fmt.Sprintf("Client '%v' not found. Add into DB", ticketDetails.CustomerID))
		err := clientModule.AddClient(ticketDetails.CustomerID)
		if err != nil {
			// TODO - add logic for close program
			p.Log.Error(fmt.Sprintf("Can't add new client '%v'", ticketDetails.CustomerID))
			return
		}
	case myErrors.ErrMoreThenOneTeamBounded:
		p.Log.Error(fmt.Sprintf("With client '%v' bound more than one team. Send message to all users.", ticketDetails.CustomerID))
		go sendMessageForAllBotUsers(message, p.DB, p.Log, p.Telegram)
	default:
		p.Log.Error(fmt.Sprintf("While get bounded team for client '%v'", ticketDetails.CustomerID))
		// TODO - add logic for close program
		return
	}

	// Suspend event processing.
	err = (*p.DB).OTRSEventSuspend(eventDBID)
	if err != nil {
		p.Log.Debug(fmt.Sprintf("Can't suspend event with ID '%v' - '%v'", eventDBID, err))
		// TODO - add logic for close program
	}
}

func sendMessageForAllBotUsers(message string, db *DBProvider.DBProvider, logger logger.Logger, tBot *TelegramProvider.TelegramProvider) {
	logger.Debug(fmt.Sprintf("Start sending sequense for message:\n'%v'", message))

	// Get all users by subscription (all subscriptions).
	subscriptionList := make([]string, 0, 8)
	subscriptionList = append(subscriptionList, "Team1", "Team2", "Team3")
	userList, err := (*db).SubscriptionListGetActiveByMultipleSubscription(subscriptionList)
	if err != nil {
		logger.Error(fmt.Sprintf("Whle get users by subscription - '%v'", err))
		return
	}
	if len(userList) < 1 {
		logger.Warning(fmt.Sprintf("No users for all subscriptions"))
		return
	}

	// Generate send message tasks.
	for _, user := range userList {
		go sendMessage(user, &message, db, tBot, logger)
	}
}

func sendMessageForByTeam(team, message string, db *DBProvider.DBProvider, logger logger.Logger, tBot *TelegramProvider.TelegramProvider) {
	logger.Debug(fmt.Sprintf("Start sending sequense for message:\n'%v'", message))

	// Get all users by subscription.
	userList, err := (*db).SubscriptionListGetActiveBySubscription(team)
	if err != nil {
		logger.Error(fmt.Sprintf("Whle get users by subscription - '%v'", err))
		return
	}
	if len(userList) < 1 {
		logger.Warning(fmt.Sprintf("No users for '%v' subscription", team))
		return
	}

	// Generate send message tasks.
	for _, user := range userList {
		go sendMessage(user, &message, db, tBot, logger)
	}
}

func sendMessage(userID int64, message *string, db *DBProvider.DBProvider, tBot *TelegramProvider.TelegramProvider, logger logger.Logger) {
	logger.Debug(fmt.Sprintf("Start sending message to telegram user '%v'", userID))

	// Get user telegram ID
	telegramID, err := (*db).BotUserGetTelegramIDByID(userID)
	if err != nil {
		logger.Error(fmt.Sprintf("While get user's telegram ID - '%v'. Message not sent or scheduled.", err))
		return
	}

	// Schedule message.
	messageID, err := (*db).MessageListNewMessage("Telegram", telegramID, *message)
	if err != nil {
		logger.Error(fmt.Sprintf("While scheduling message - '%v'. Message not sent or scheduled.", err))
		return
	}

	// Send message into social media.
	err = (*tBot).SendEventMessage(userID, *message)
	if err != nil {
		logger.Error(fmt.Sprintf("While send message - '%v'. Try again in 1 minute.", err))
		err = (*tBot).SendEventMessage(userID, *message)
		if err != nil {
			logger.Error(fmt.Sprintf("While retry send message - '%v'. Message not sent. Retry in next retry for all failed messages", err))
			err = (*tBot).SendEventMessage(userID, *message)
			return
		}
	}

	// Finish message processing.
	logger.Debug(fmt.Sprintf("Message to telegram user '%v' sucessfully sent", userID))
	err = (*db).MessageListMarkDelivered(messageID)
	if err != nil {
		logger.Error(fmt.Sprintf("While mark message as delivered - '%v'. Message can be sent twice.", err))
		return
	}
}

func finishEventProcessing(reason string, eventID int64, db *DBProvider.DBProvider, logger logger.Logger) {
	logger.Debug(fmt.Sprintf("Event with ID '%v' finished. Reason '%v'", eventID, reason))

	// Set event as finished.
	err := (*db).OTRSEventEnded(eventID)
	if err != nil {
		logger.Debug(fmt.Sprintf("Can't finish event with ID '%v' and reason '%v' - '%v'", eventID, reason, err))
		// TODO - add logic for close program
	}
}
