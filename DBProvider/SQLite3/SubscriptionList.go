package SQLite3

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"time"
)

// Return all active subscriptions for provided user.
// If user have no subscriptions return empty slice.
func (db *DB) SubscriptionListGetActiveByUser(userID uint64) ([]string, error) {
	// Create new sql transaction.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for scan subscriptions for user '%v' - '%v'", userID, err))
		return nil, err
	}

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
		return nil, err
	}

	return subscriptionList, nil
}

func (db *DB) SubscriptionListAdd(userID uint64, newSubscription string) error {
	// Check if user already subscribed and return ErrAlreadySubscribed if so.
	activeSubscriptionList, err := db.SubscriptionListGetActiveByUser(userID)
	if err != nil {
		return err
	}
	for _, activeSubscription := range activeSubscriptionList {
		if activeSubscription == newSubscription {
			return errors.ErrAlreadySubscribed
		}
	}

	// Create new sql transaction for add new subscription.
	transaction, err := db.Instance.Begin()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Can't create transaction for add subscription '%v' for user '%v' - '%v'",
			newSubscription, userID, err))
		return err
	}

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
		return err
	}

	// Close transaction.
	err = transaction.Commit()
	if err != nil {
		db.Log.Error(fmt.Sprintf("Wile commit transaction transaction for add subscription '%v' for user '%v' - '%v'",
			newSubscription, userID, err))
		return err
	}

	return nil
}
