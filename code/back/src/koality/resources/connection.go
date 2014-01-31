package resources

type Closer interface {
	Close() error
}

type Connection struct {
	Users          *UsersHandler
	Repositories   *RepositoriesHandler
	Verifications  *VerificationsHandler
	Stages         *StagesHandler
	Pools          *PoolsHandler
	Settings       *SettingsHandler
	Snapshots      *SnapshotsHandler
	DebugInstances *DebugInstancesHandler
	Closer
}
