package infile

import (
	"github.com/Sarraksh/otrs-echo-bot/common/errors"
	"sync"
)

type InFile struct {
	UserData         map[uint64]User
	LastUserID       uint64
	SubscriptionData SubscriptionData
	mx               sync.Mutex
	NotSaved         bool
}

// TODO - Initialise()
func (i *InFile) Initialise() error {
	panic("implement me")
}

// TODO - Load()
func (i *InFile) Load() error {
	panic("implement me")
}

// TODO - Save()
func (i *InFile) Save() error {
	panic("implement me")
}

// Create new user and return his ID
func (i *InFile) AddUser() uint64 {
	i.mx.Lock()
	defer i.mx.Unlock()
	i.LastUserID++
	i.UserData[i.LastUserID] = User{
		FirstName:           "",
		LastName:            "",
		SocialMediaDataList: make(map[string]SocialMediaData, 8),
	}
	return i.LastUserID
}

func (i *InFile) AddUserSMID(ID uint64, socialMediaName, socialMediaID string) {
	i.mx.Lock()
	socialMediaData := i.UserData[ID].SocialMediaDataList[socialMediaName]
	socialMediaData.ID = socialMediaID
	i.UserData[ID].SocialMediaDataList[socialMediaName] = socialMediaData
	i.mx.Unlock()
}

func (i *InFile) GetUserBySMID(socialMediaID string) (uint64, error) {
	i.mx.Lock()
	defer i.mx.Unlock()
	for id, user := range i.UserData {
		for _, SMData := range user.SocialMediaDataList {
			if SMData.ID == socialMediaID {
				return id, nil
			}
		}
	}
	return 0, errors.ErrUserNotFound
}

func (i *InFile) AddSub(userID uint64, subName string) error {
	i.mx.Lock()
	defer i.mx.Unlock()
	subList := (*i).SubscriptionData[subName]
	err := subList.add(userID)
	if err != nil {
		return err
	}
	(*i).SubscriptionData[subName] = subList
	return nil
}

func (i *InFile) RemoveSub(userID uint64, subName string) error {
	i.mx.Lock()
	defer i.mx.Unlock()
	subList := (*i).SubscriptionData[subName]
	err := subList.remove(userID)
	if err != nil {
		return err
	}
	(*i).SubscriptionData[subName] = subList
	return nil
}
