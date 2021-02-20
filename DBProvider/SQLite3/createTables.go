package SQLite3

import (
	"database/sql"
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
)

const (
	sqlCreateBotUserListTable = `
create table BotUserList (
	ID integer not null primary key,
	Token text not null,
	Active integer not null,
	FirstName text,
	LastName text,
	Phone integer,
	Email text,
	Created integer not null,
	TelegramID integer
);`
	sqlCreateOTRSEventListTable = `
create table OTRSEventList (
	ID integer not null primary key,
	Status text not null,
	Channel text not null,
	Type text not null,
	TicketID integer not null,
	Created integer not null,
	ActivationInterval integer,
	NextActivation integer,
	Finished integer
);`
	sqlCreateSubscriptionListTable = `
create table SubscriptionList (
	Active integer not null,
	Subscription text not null,
	UserID integer not null,
	Created integer not null,
	Finished integer,
	PRIMARY KEY (UserID, Subscription)
);`
	sqlCreateSubscriptionSchedulerTable = `
create table SubscriptionScheduler (
	Active integer not null,
	Subscription text not null,
	UserID integer not null,
	CreateIn integer not null,
	DeleteIn integer,
	PRIMARY KEY (UserID, Subscription)
);`
	sqlCreateMessageListTable = `
create table MessageList (
	ID integer not null primary key,
	SocialMedia text not null,
	ChatID text not null,
	Text text not null,
	Created integer not null,
	Sent integer
);`
	sqlCreateClientsAssignedToTeamsTable = `
create table MessageList (
	ID integer not null primary key,
	SocialMedia text not null,
	ChatID text not null,
	Text text not null,
	Created integer not null,
	Sent integer
);`
)

// If tables don't exist create them.
func createAllTablesIfNotExist(db *sql.DB, Log logger.Logger) error {
	Log.Debug("Start createAllTablesIfNotExist sequence")

	tableCreateStatementList := make(map[string]string)
	tableCreateStatementList["BotUserList"] = sqlCreateBotUserListTable
	tableCreateStatementList["OTRSEventList"] = sqlCreateOTRSEventListTable
	tableCreateStatementList["SubscriptionList"] = sqlCreateSubscriptionListTable
	tableCreateStatementList["SubscriptionScheduler"] = sqlCreateSubscriptionSchedulerTable
	tableCreateStatementList["MessageList"] = sqlCreateMessageListTable
	tableCreateStatementList["ClientsAssignedToTeams"] = sqlCreateClientsAssignedToTeamsTable

	for currentTable, sqlStatement := range tableCreateStatementList {
		Log.Debug(fmt.Sprintf("Processing '%+v' table", currentTable))
		tableExist, err := isTableExists(db, Log, currentTable)
		if err != nil {
			return err
		}
		if tableExist {
			continue
		}
		Log.Debug(fmt.Sprintf("Table '%+v' not exist. Create it", currentTable))
		_, err = db.Exec(sqlStatement)
		if err != nil {
			return err
		}
	}

	Log.Debug("Sequence createAllTablesIfNotExist successfully finished")
	return nil
}
