package SQLite3

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/myErrors"
	"time"
)

// Return all active subscriptions for provided user.
// If user have no subscriptions return empty slice.
func (db *DB) SubscriptionListGetActiveByUser(userID int64) ([]string, error) {
	db.Log.Debug(fmt.Sprintf("Collect subscription list by user '%+v'", userID))
	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for scan subscriptions for user '%v' - '%v'", userID, err))
		return nil, err
	}
	defer transaction.Rollback()

	// Prepare transaction for select from table.
	statement, err := transaction.Prepare(`SELECT Subscription FROM SubscriptionList WHERE UserID = ? AND Active = 1;`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for scan subscriptions for user '%v' - '%v'", userID, err))
		return nil, err
	}
	defer statement.Close()

	// Query active subscription list for user.
	rows, err := statement.Query(userID)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't query for scan subscriptions for user '%v' - '%v'", userID, err))
		return nil, err
	}
	defer rows.Close()

	// Check query result.
	var subscriptionList = make([]string, 0, 16)
	var subscription string
	for rows.Next() {
		err = rows.Scan(&subscription)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan subscriptions for user '%v' - '%v'", userID, err))
			return nil, err
		}
		subscriptionList = append(subscriptionList, subscription)
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for subscriptions for user '%v' - '%v'", userID, err))
		return nil, err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for subscriptions for user '%v' - '%v'", userID, err))
		return nil, err
	}

	db.Log.Debug(fmt.Sprintf("Collected subscription list by user '%+v'", userID))
	return subscriptionList, nil
}

// Return list of users with active provided subscription name.
func (db *DB) SubscriptionListGetActiveBySubscription(subscription string) ([]int64, error) {
	db.Log.Debug(fmt.Sprintf("Collect users by subscription '%+v'", subscription))
	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for scan users by subscription '%v' - '%v'", subscription, err))
		return nil, err
	}
	defer transaction.Rollback()

	// Prepare transaction for select from table.
	statement, err := transaction.Prepare(`SELECT UserID FROM SubscriptionList WHERE Subscription = ? AND Active = 1;`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for scan users by subscription '%v' - '%v'", subscription, err))
		return nil, err
	}
	defer statement.Close()

	// Query active users list for subscription.
	rows, err := statement.Query(subscription)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't query for scan users by subscription '%v' - '%v'", subscription, err))
		return nil, err
	}
	defer rows.Close()

	// Check query result.
	var userList = make([]int64, 0, 32)
	var user int64
	for rows.Next() {
		err = rows.Scan(&user)
		if err != nil {
			db.Log.Error(fmt.Sprintf("Can't scan users by subscription '%v' - '%v'", subscription, err))
			return nil, err
		}
		userList = append(userList, user)
	}
	err = rows.Err()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While iteration for users by subscription '%v' - '%v'", subscription, err))
		return nil, err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("While commit for users by subscription '%v' - '%v'", subscription, err))
		return nil, err
	}

	db.Log.Debug(fmt.Sprintf("Collected users '%+v'", userList))
	return userList, nil
}

// Collect all users for all subscriptions and remove duplicates.
func (db *DB) SubscriptionListGetActiveByMultipleSubscription(subscriptionList []string) ([]int64, error) {
	db.Log.Debug(fmt.Sprintf("Collect users by subscriptions '%+v'", subscriptionList))
	// Collect all users for all subscriptions.
	var userList = make([]int64, 0, 128)
	for _, subscription := range subscriptionList {
		tmpUserList, err := db.SubscriptionListGetActiveBySubscription(subscription)
		if err != nil {
			return nil, err
		}
		userList = append(userList, tmpUserList...)
	}

	// Remove duplicate users.
	var presentMap = make(map[int64]bool)
	userCount := len(userList)
	for i := 0; i < userCount; {
		if presentMap[userList[i]] {
			userList[i] = userList[userCount-1]
			userCount--
		} else {
			presentMap[userList[i]] = true
			i++
		}
	}
	userList = userList[:userCount]

	db.Log.Debug(fmt.Sprintf("Collected users '%+v'", userList))
	return userList, nil
}

// Add new subscription for user if not already subscribed.
func (db *DB) SubscriptionListAdd(userID int64, newSubscription string) error {
	db.Log.Debug(fmt.Sprintf("Add subscription '%+v' for user '%+v'", newSubscription, userID))
	// Check if user already subscribed and return ErrAlreadySubscribed if so.
	activeSubscriptionList, err := db.SubscriptionListGetActiveByUser(userID)
	if err != nil {
		return err
	}
	for _, activeSubscription := range activeSubscriptionList {
		if activeSubscription == newSubscription {
			db.Log.Debug(fmt.Sprintf("User already subscribed for '%v'", newSubscription))
			return myErrors.ErrAlreadySubscribed
		}
	}

	// Create new sql transaction for add new subscription.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for add subscription '%v' for user '%v' - '%v'",
			newSubscription, userID, err))
		return err
	}
	defer transaction.Rollback()

	// Prepare transaction for add new subscription.
	statement, err := transaction.Prepare(`INSERT INTO SubscriptionList(Active, Subscription, UserID, Created) VALUES(?, ?, ?, ?);`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for add subscription '%v' for user '%v' - '%v'",
			newSubscription, userID, err))
		return err
	}
	defer statement.Close()

	// Execute prepared statement
	_, err = statement.Exec(1, newSubscription, userID, time.Now().Unix())
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't execute statement for add subscription '%v' for user '%v' - '%v'",
			newSubscription, userID, err))
		return err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Wile commit transaction for add subscription '%v' for user '%v' - '%v'",
			newSubscription, userID, err))
		return err
	}

	db.Log.Debug(fmt.Sprintf("Succesful add subscription '%+v' for user '%+v'", newSubscription, userID))
	return nil
}

// Unsubscribe user if he currently subscribed.
func (db *DB) SubscriptionListRemove(userID int64, removeSubscription string) error {
	db.Log.Debug(fmt.Sprintf("Remove subscription '%+v' for user '%+v'", removeSubscription, userID))
	// Check if user subscribed. If not, return ErrNotSubscribed.
	activeSubscriptionList, err := db.SubscriptionListGetActiveByUser(userID)
	if err != nil {
		return err
	}
	var subscribed bool = false
	for _, activeSubscription := range activeSubscriptionList {
		if activeSubscription == removeSubscription {
			subscribed = true
		}
	}
	if !subscribed {
		return myErrors.ErrNotSubscribed
	}

	// Create new sql transaction for remove subscription.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for remove subscription '%v' for user '%v' - '%v'",
			removeSubscription, userID, err))
		return err
	}
	defer transaction.Rollback()

	// Prepare transaction for remove subscription.
	statement, err := transaction.Prepare(`UPDATE SubscriptionList SET Active = 0, Finished = ?
WHERE Active = 1 AND Subscription = ? AND UserID = ?;`)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't prepare transaction for remove subscription '%v' for user '%v' - '%v'",
			removeSubscription, userID, err))
		return err
	}
	defer statement.Close()

	// Execute prepared statement
	_, err = statement.Exec(time.Now().Unix(), removeSubscription, userID)
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't execute statement for remove subscription '%v' for user '%v' - '%v'",
			removeSubscription, userID, err))
		return err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Wile commit transaction for remove subscription '%v' for user '%v' - '%v'",
			removeSubscription, userID, err))
		return err
	}

	db.Log.Debug(fmt.Sprintf("Succesful remove subscription '%+v' for user '%+v'", removeSubscription, userID))
	return nil
}
