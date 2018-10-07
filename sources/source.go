package sources

// Service represents the methods to backup/restore a service
type Source interface {
	Backup() (string, error)
	Restore(path string) error
}
