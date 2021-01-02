package storage

// Store users and subscriptions data and synchronise it with persistent storage.
type Storage interface {
	// Initialise persistent storage.
	// If any data already present return ErrPersistentStorageNotEmpty.
	Initialise() error

	// Load existing data if present in persistent storage.
	// If storage empty return ErrPersistentStorageEmpty.
	Load() error

	//Save data into persistent storage.
	Save() error

	// Create new empty user and return his ID.
	AddUser() uint64

	AddUserSMID(ID uint64, socialMediaName, socialMediaID string)

	// Get user ID by his social media ID.
	GetUserBySMID(socialMediaID string) uint64

	// Add user ID to specified subscribers list.
	AddSub(userID uint64, subName string) error

	// Remove user ID from specified subscribers list.
	RemoveSub(userID uint64, subName string) error
}
