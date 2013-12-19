package resources

type Connection struct {
	Users         *UsersHandler
	Repositories  *RepositoriesHandler
	Verifications *VerificationsHandler
	Stages        *StagesHandler
	Pools         *PoolsHandler
}
