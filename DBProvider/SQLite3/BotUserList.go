package SQLite3

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"time"
)

//
func (db *DB) BotUserAdd(tgID int64) error {
	db.Log.Info(fmt.Sprintf("Write new user with telegram ID '%+v'", tgID))

	// Prepare data for insert.
	db.LastIDmx.Lock()
	userID := db.LastID.BotUserList + 1
	db.LastID.BotUserList = userID
	db.LastIDmx.Unlock()
	Token := "" // TODO - implement token generation and fill token for all old users
	Active := 1 // User is active
	FirstName := ""
	LastName := ""
	Phone := 0
	Email := ""
	Created := time.Now().Unix()

	// Prepare insert string.
	Insert := "insert into BotUserList(ID, Token, Active, FirstName, TicketID, LastName, Phone, Email, Created, TelegramID) values"
	sqlStatement := fmt.Sprintf(
		"%s(%d, '%s', %d, '%s', '%s', %d, '%s', %d, %d)",
		Insert,
		userID,
		Token,
		Active,
		FirstName,
		LastName,
		Phone,
		Email,
		Created,
		tgID,
	)
	db.Log.Debug(fmt.Sprintf("Insert string - '%+v'", sqlStatement))

	// Write data into DB.
	_, err := db.Instance.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

// Change user first name. Find  user by telegram ID.
func (db *DB) BotUserUpdateFirstName(tgID int64, firstName string) error {
	// Search for user ID.
	userID, err := db.BotUserGetByTelegramID(tgID)
	if err != nil {
		return err
	}

	// Update data into DB.
	sqlStatement := fmt.Sprintf("UPDATE BotUserList SET FirstName = '%s' WHERE ID = %d;", firstName, userID)
	db.Log.Debug(fmt.Sprintf("Query string '%v'", sqlStatement))
	_, err = db.Instance.Exec(sqlStatement)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get update BotUserList - '%+v'", err))
		return err
	}

	return nil
}

// Change user last name. Find  user by telegram ID.
func (db *DB) BotUserUpdateLastName(tgID int64, lastName string) error {
	// Search for user ID.
	userID, err := db.BotUserGetByTelegramID(tgID)
	if err != nil {
		return err
	}

	// Update data into DB.
	sqlStatement := fmt.Sprintf("UPDATE BotUserList SET LastName = '%s' WHERE ID = %d;", lastName, userID)
	db.Log.Debug(fmt.Sprintf("Query string '%v'", sqlStatement))
	_, err = db.Instance.Exec(sqlStatement)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get update BotUserList - '%+v'", err))
		return err
	}

	return nil
}

func (db *DB) BotUserGetByTelegramID(tgID int64) (int64, error) {
	// Query table BotUser for user ID.
	queryString := fmt.Sprintf("SELECT ID FROM BotUserList WHERE Status TelegramID = %d;", tgID)
	db.Log.Debug(fmt.Sprintf("Query string '%v'", queryString))
	rows, err := db.Instance.Query(queryString)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get bot user by telegram ID - '%+v'", err))
		return 0, err
	}
	defer rows.Close()

	// Check query result
	var userID int64 = 0
	numberOfUsers := 0
	for rows.Next() {
		numberOfUsers++ // Count received rows.
		err = rows.Scan(&userID)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan for bot user by telegram ID - '%+v'", err))
			return 0, err
		}
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for bot user by telegram ID '%v'", err))
		return 0, err
	}

	// Check if more than one row or no raws received.
	switch {
	case numberOfUsers > 1:
		return userID, errors.ErrMoreThanOneUser
	case numberOfUsers == 0:
		return 0, errors.ErrNoUsersFound
	}

	return userID, nil
}
