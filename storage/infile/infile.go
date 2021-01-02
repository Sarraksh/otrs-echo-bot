package infile

type InFile struct {
	UserData      map[uint64]User
	LastUserID    uint64
	Subscriptions Subscriptions
}

type User struct {
	SocialMediaID map[string]string
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

func (i InFile) AddUserSMID(ID uint64, socialMediaID string) {
	panic("implement me")
}

func (i InFile) GetUserBySMID(socialMediaID string) uint64 {
	panic("implement me")
}

func (i InFile) AddSub(userID uint64, subName string) error {
	panic("implement me")
}

func (i InFile) RemoveSub(userID uint64, subName string) error {
	panic("implement me")
}
