package infile

import (
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"time"
)

// Store and manage subscribers for each subscription.
type SubscriberList []Subscriber

//
func (s *SubscriberList) add(newSub uint64) error {
	currentSub := Subscriber{
		UserID:           newSub,
		SubscriptionTime: time.Now(),
	}
	if len(*s) == 0 {
		*s = make([]Subscriber, 0, 32)
		*s = append(*s, currentSub)
		return nil
	}
	for _, sub := range *s {
		if sub.UserID == newSub {
			return errors.ErrAlreadySubscribed
		}
	}
	*s = append(*s, currentSub)
	return nil
}

func (s *SubscriberList) remove(remSub uint64) error {
	if len(*s) == 0 {
		return errors.ErrNotSubscribed
	}
	for i, sub := range *s {
		if sub.UserID == remSub {
			(*s)[i] = (*s)[len(*s)-1]
			*s = (*s)[:len(*s)-1]
			return nil
		}
	}
	return errors.ErrNotSubscribed
}
