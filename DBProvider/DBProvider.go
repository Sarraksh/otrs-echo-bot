package DBProvider

import "github.com/Sarraksh/otrs-echo-bot/common/logger"

type DBProvider interface {
	Initialise(logger logger.Logger, directory string) error

	OTRSEventCreateNew(Channel, Type, TicketID string) error
	OTRSEventGetActive() (int64, string, error)
	OTRSEventGetEarliestActivationTimestamp() (int64, error)
	OTRSEventSuspend(id int64) error
	OTRSEventEnded(id int64) error
	OTRSEventIsExistsWithTicketIDAndType(ticketID int64, eventType string) (bool, error)

	BotUserAdd(tgID int64) error
	BotUserUpdateFirstName(tgID int64, firstName string) error
	BotUserUpdateLastName(tgID int64, lastName string) error
	BotUserGetByTelegramID(tgID int64) (int64, error)

	SubscriptionListGetActiveByUser(userID int64) ([]string, error)
	SubscriptionListGetActiveBySubscription(subscription string) ([]int64, error)
	SubscriptionListGetActiveByMultipleSubscription(subscriptionList []string) ([]int64, error)
	SubscriptionListAdd(userID int64, newSubscription string) error
	SubscriptionListRemove(userID int64, removeSubscription string) error

	ClientTeamBoundClientAdd(client, team string) error
	ClientTeamBoundClientUpdate(client, team string) error
	ClientTeamBoundGetTeamByClient(client string) (string, error)

	MessageListNewMessage(sm, chatID, text string) (int64, error)
	MessageListMarkDelivered(ID int64) error
	MessageListGetAllUndeliveredBySM(sm string) ([]int64, error)
	MessageListGetMessageText(ID int64) (string, error)
	MessageListGetMessageChatID(ID int64) (string, error)
}
