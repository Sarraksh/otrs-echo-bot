package DBProvider

import "github.com/Sarraksh/otrs-echo-bot/common/logger"

type DBProvider interface {
	Initialise(logger logger.Logger, directory string) error

	OTRSEventCreateNew(Channel, Type, TicketID string) error
	OTRSEventGetActive() (int64, string, error)
	OTRSEventGetEarliestActivationTimestamp() (int64, error)
	OTRSEventSuspend(id int64) error
	OTRSEventEnded(id int64) error

	BotUserAdd(tgID int64) error
	BotUserUpdateFirstName(tgID int64, firstName string) error
	BotUserUpdateLastName(tgID int64, lastName string) error
	BotUserGetByTelegramID(tgID int64) (int64, error)

	SubscriptionListGetActiveByUser(userID uint64) ([]string, error)
	SubscriptionListGetActiveBySubscription(subscription string) ([]uint64, error)
	SubscriptionListGetActiveByMultipleSubscription(subscriptionList []string) ([]uint64, error)
	SubscriptionListAdd(userID uint64, newSubscription string) error
	SubscriptionListRemove(userID uint64, removeSubscription string) error
}
