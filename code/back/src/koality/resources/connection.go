package resources

type Closer interface {
	Close() error
}

type Connection struct {
	Users          *UsersHandler
	Repositories   *RepositoriesHandler
	Builds         *BuildsHandler
	Stages         *StagesHandler
	Pools          *PoolsHandler
	Settings       *SettingsHandler
	Snapshots      *SnapshotsHandler
	DebugInstances *DebugInstancesHandler
	Closer
}
