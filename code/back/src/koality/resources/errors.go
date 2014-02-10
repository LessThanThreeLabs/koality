package resources

//-------------------- Users --------------------

type UserAlreadyExistsError struct {
	Message string
}

func (err UserAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchUserError struct {
	Message string
}

func (err NoSuchUserError) Error() string {
	return err.Message
}

type KeyAlreadyExistsError struct {
	Message string
}

func (err KeyAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchKeyError struct {
	Message string
}

func (err NoSuchKeyError) Error() string {
	return err.Message
}

//-------------------- Repositories --------------------

type RepositoryAlreadyExistsError struct {
	Message string
}

func (err RepositoryAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchRepositoryError struct {
	Message string
}

func (err NoSuchRepositoryError) Error() string {
	return err.Message
}

type InvalidRepositoryRemoteUriError struct {
	Message string
}

func (err InvalidRepositoryRemoteUriError) Error() string {
	return err.Message
}

type InvalidRepositoryStatusError struct {
	Message string
}

func (err InvalidRepositoryStatusError) Error() string {
	return err.Message
}

type NoSuchRepositoryHookError struct {
	Message string
}

func (err NoSuchRepositoryHookError) Error() string {
	return err.Message
}

type InvalidRepositoryHookTypeError struct {
	Message string
}

func (err InvalidRepositoryHookTypeError) Error() string {
	return err.Message
}

//-------------------- Pools --------------------

type NoSuchPoolError struct {
	Message string
}

func (err NoSuchPoolError) Error() string {
	return err.Message
}

type PoolAlreadyExistsError struct {
	Message string
}

func (err PoolAlreadyExistsError) Error() string {
	return err.Message
}

//-------------------- Snapshots --------------------

type NoSuchSnapshotError struct {
	Message string
}

func (err NoSuchSnapshotError) Error() string {
	return err.Message
}

type InvalidSnapshotStatusError struct {
	Message string
}

func (err InvalidSnapshotStatusError) Error() string {
	return err.Message
}

//-------------------- Debug Instances --------------------

type NoSuchDebugInstanceError struct {
	Message string
}

func (err NoSuchDebugInstanceError) Error() string {
	return err.Message
}

//-------------------- Builds --------------------

type NoSuchBuildError struct {
	Message string
}

func (err NoSuchBuildError) Error() string {
	return err.Message
}

type InvalidBuildStatusError struct {
	Message string
}

func (err InvalidBuildStatusError) Error() string {
	return err.Message
}

type InvalidBuildMergeStatusError struct {
	Message string
}

func (err InvalidBuildMergeStatusError) Error() string {
	return err.Message
}

type ChangesetAlreadyExistsError struct {
	Message string
}

func (err ChangesetAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchChangesetError struct {
	Message string
}

func (err NoSuchChangesetError) Error() string {
	return err.Message
}

//-------------------- Stages --------------------

type NoSuchStageError struct {
	Message string
}

func (err NoSuchStageError) Error() string {
	return err.Message
}

type StageAlreadyExistsError struct {
	Message string
}

func (err StageAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchStageRunError struct {
	Message string
}

func (err NoSuchStageRunError) Error() string {
	return err.Message
}

type NoSuchXunitError struct {
	Message string
}

func (err NoSuchXunitError) Error() string {
	return err.Message
}

//-------------------- Settings --------------------

type NoSuchSettingError struct {
	Message string
}

func (err NoSuchSettingError) Error() string {
	return err.Message
}
