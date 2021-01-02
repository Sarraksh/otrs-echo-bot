package infile

import "sync"

type InFile struct {
	UserData         map[uint64]User
	LastUserID       uint64
	SubscriptionData SubscriptionData
	mx               sync.Mutex
}

func (i InFile) Initialise() error {
	panic("implement me")
}

func (i InFile) Load() error {
	panic("implement me")
}

func (i InFile) Save() error {
	panic("implement me")
}

func (i InFile) AddUser() uint64 {
	panic("implement me")
}

func (i InFile) AddUserSMID(ID uint64, socialMediaName, socialMediaID string) {
	panic("implement me")
}

func (i InFile) GetUserBySMID(socialMediaID string) uint64 {
	panic("implement me")
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
