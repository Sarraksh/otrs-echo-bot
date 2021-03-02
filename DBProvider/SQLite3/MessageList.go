package SQLite3

import (
	"fmt"
	"time"
)

// Add new message. Return message ID in DB.
func (db *DB) MessageListNewMessage(sm, chatID, text string) (int64, error) {
	db.Log.Debug(fmt.Sprintf("Add new message for chat '%v' in '%v'. Text - '%v'", chatID, sm, text))

	// Prepare data for insert.
	db.LastIDmx.Lock()
	ID := db.LastID.MessageList + 1
	db.LastID.MessageList = ID
	db.LastIDmx.Unlock()
	Created := time.Now().Unix()

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for add new message for caht '%v' in '%v' - '%v'", chatID, sm, err))
		return 0, err
	}
	defer transaction.Rollback()

	// Prepare transaction for insert into table.
	statement, err := transaction.Prepare(`INSERT INTO MessageList(ID, SocialMedia, ChatID, MessageText, Created) VALUES(?, ?, ?, ?, ?);`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for add new message for caht '%v' in '%v' - '%v'", chatID, sm, err))
		return 0, err
	}
	defer statement.Close()

	// Execute statement.
	_, err = statement.Exec(ID, sm, chatID, text, Created)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't execute transaction for add new message for caht '%v' in '%v' - '%v'", chatID, sm, err))
		return 0, err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for add new message for caht '%v' in '%v' - '%v'", chatID, sm, err))
		return 0, err
	}

	db.Log.Debug(fmt.Sprintf("Succesfully add new message for chat '%v' in '%v'. Text - '%v'", chatID, sm, text))
	return ID, nil
}

// Mark message as delivered dy message ID.
func (db *DB) MessageListMarkDelivered(ID int64) error {
	db.Log.Debug(fmt.Sprintf("Mark message ID '%v' as delivered", ID))

	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for mark message ID '%v' as delivered", ID))
		return err
	}
	defer transaction.Rollback()

	// Prepare transaction for update table.
	statement, err := transaction.Prepare(`UPDATE MessageList SET Sent = ? WHERE ID = ?`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for mark message ID '%v' as delivered", ID))
		return err
	}
	defer statement.Close()

	// Execute statement.
	_, err = statement.Exec(ID)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't execute transaction for mark message ID '%v' as delivered", ID))
		return err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for mark message ID '%v' as delivered", ID))
		return err
	}

	db.Log.Debug(fmt.Sprintf("Succesfully mark message ID '%v' as delivered", ID))
	return nil
}

// Get all messages undelivered to social media API.
func (db *DB) MessageListGetAllUndeliveredBySM(sm string) ([]int64, error) {
	db.Log.Debug(fmt.Sprintf("Get all undelivered messages for '%+v'", sm))
	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get all undelivered messages for '%v' - '%v'", sm, err))
		return nil, err
	}
	defer transaction.Rollback()

	// Prepare transaction for select from table.
	statement, err := transaction.Prepare(`SELECT ID FROM MessageList WHERE Sent < ?;`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for get all undelivered messages for '%v' - '%v'", sm, err))
		return nil, err
	}
	defer statement.Close()

	// Leave time gap for just send messages.
	timeThreshold := time.Now().Unix() + 60

	// Query message list.
	rows, err := statement.Query(timeThreshold)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't query for get all undelivered messages for '%v' - '%v'", sm, err))
		return nil, err
	}
	defer rows.Close()

	// Check query result.
	var messageID int64 = 0
	var messageIDList = make([]int64, 0, 128)
	for rows.Next() {
		err = rows.Scan(&messageID)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan undelivered messages for '%v' - '%v'", sm, err))
			return nil, err
		}
		messageIDList = append(messageIDList, messageID)
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for get all undelivered messages for '%v' - '%v'", sm, err))
		return nil, err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for get all undelivered messages for '%v' - '%v'", sm, err))
		return nil, err
	}

	db.Log.Debug(fmt.Sprintf("Sucessful all undelivered messages for '%+v'", sm))
	return messageIDList, nil
}

// Get message text by message ID.
func (db *DB) MessageListGetMessageText(ID int64) (string, error) {
	db.Log.Debug(fmt.Sprintf("Get message text by message ID '%+v'", ID))
	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get message text by message ID '%+v'", ID))
		return "", err
	}
	defer transaction.Rollback()

	// Prepare transaction for select from table.
	statement, err := transaction.Prepare(`SELECT MessageText FROM MessageList WHERE ID = ?;`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for message text by message ID '%+v'", ID))
		return "", err
	}
	defer statement.Close()

	// Leave time gap for just send messages.
	timeThreshold := time.Now().Unix() + 60

	// Query message text.
	rows, err := statement.Query(timeThreshold)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't query for get message text by message ID '%+v'", ID))
		return "", err
	}
	defer rows.Close()

	// Check query result.
	text := ""
	for rows.Next() {
		err = rows.Scan(&text)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan message text by message ID '%+v'", ID))
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for get message text by message ID '%+v'", ID))
		return "", err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for get message text by message ID '%+v'", ID))
		return "", err
	}

	db.Log.Debug(fmt.Sprintf("Sucessful get message text by message ID '%+v'", ID))
	return text, nil
}

// Get chat ID for message by message ID.
func (db *DB) MessageListGetMessageChatID(ID int64) (string, error) {
	db.Log.Debug(fmt.Sprintf("Get chat ID by message ID '%+v'", ID))
	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't get chat ID by message ID '%+v'", ID))
		return "", err
	}
	defer transaction.Rollback()

	// Prepare transaction for select from table.
	statement, err := transaction.Prepare(`SELECT ChatID FROM MessageList WHERE ID = ?;`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for chat ID by message ID '%+v'", ID))
		return "", err
	}
	defer statement.Close()

	// Leave time gap for just send messages.
	timeThreshold := time.Now().Unix() + 60

	// Query chat ID.
	rows, err := statement.Query(timeThreshold)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't query for get chat ID by message ID '%+v'", ID))
		return "", err
	}
	defer rows.Close()

	// Check query result.
	chatID := ""
	for rows.Next() {
		err = rows.Scan(&chatID)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan chat ID by message ID '%+v'", ID))
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for get chat ID by message ID '%+v'", ID))
		return "", err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for get chat ID by message ID '%+v'", ID))
		return "", err
	}

	db.Log.Debug(fmt.Sprintf("Sucessful get chat ID by message ID '%+v'", ID))
	return chatID, nil
}
