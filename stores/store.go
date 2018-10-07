package stores

// Storer represents the methods to store/retrieve a backup from another location
type Store interface {
	Store(filepath string, filename string) error
	Retrieve(s3path string) (string, error)
	RemoveOlderBackups(keep int) error
	FindLatestBackup() (string, error)
	Close()
}
