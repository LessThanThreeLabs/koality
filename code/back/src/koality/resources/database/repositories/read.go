package repositories

import (
	"database/sql"
	"koality/resources"
	"strings"
)

type Scannable interface {
	Scan(dest ...interface{}) error
}

type ReadHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalRepositoriesSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalRepositoriesSubscriptionHandler) (resources.RepositoriesReadHandler, error) {
	return &ReadHandler{database, verifier, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) scanRepository(scannable Scannable) (*resources.Repository, error) {
	repository := new(resources.Repository)

	var gitHubHookId sql.NullInt64
	var deletedId uint64
	var gitHubOwner, gitHubName, gitHubHookSecret, gitHubHookTypes, gitHubOAuthToken sql.NullString
	err := scannable.Scan(&repository.Id, &repository.Name, &repository.Status, &repository.VcsType,
		&repository.RemoteUri, &repository.Created, &deletedId,
		&gitHubOwner, &gitHubName, &gitHubHookId, &gitHubHookSecret, &gitHubHookTypes, &gitHubOAuthToken)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchRepositoryError{"Unable to find repository"}
	} else if err != nil {
		return nil, err
	}

	repository.IsDeleted = repository.Id == deletedId

	if gitHubOwner.Valid && gitHubName.Valid {
		repository.GitHub = new(resources.RepositoryGitHubMetadata)
	}
	if gitHubOwner.Valid {
		repository.GitHub.Owner = gitHubOwner.String
	}
	if gitHubName.Valid {
		repository.GitHub.Name = gitHubName.String
	}
	if gitHubHookId.Valid {
		repository.GitHub.HookId = gitHubHookId.Int64
	}
	if gitHubHookSecret.Valid {
		repository.GitHub.HookSecret = gitHubHookSecret.String
	}
	if gitHubHookTypes.Valid {
		repository.GitHub.HookTypes = strings.Split(gitHubHookTypes.String, ",")
	}
	if gitHubOAuthToken.Valid {
		repository.GitHub.OAuthToken = gitHubOAuthToken.String
	}
	return repository, nil
}

func (readHandler *ReadHandler) Get(repositoryId uint64) (*resources.Repository, error) {
	query := "SELECT R.id, R.name, R.status, R.vcs_type, R.remote_uri, R.created, R.deleted," +
		" RGM.owner, RGM.name, RGM.hook_id, RGM.hook_secret, RGM.hook_types, RGM.oauth_token" +
		" FROM repositories R LEFT JOIN repository_github_metadatas RGM" +
		" ON R.id=RGM.repository_id WHERE R.id=$1"
	row := readHandler.database.QueryRow(query, repositoryId)
	return readHandler.scanRepository(row)
}

func (readHandler *ReadHandler) GetByName(name string) (*resources.Repository, error) {
	query := "SELECT R.id, R.name, R.status, R.vcs_type, R.remote_uri, R.created, R.deleted," +
		" RGM.owner, RGM.name, RGM.hook_id, RGM.hook_secret, RGM.hook_types, RGM.oauth_token" +
		" FROM repositories R LEFT JOIN repository_github_metadatas RGM" +
		" ON R.id=RGM.repository_id WHERE R.name=$1"
	row := readHandler.database.QueryRow(query, name)
	return readHandler.scanRepository(row)
}

func (readHandler *ReadHandler) GetByGitHubInfo(ownerName, repositoryName string) (*resources.Repository, error) {
	query := "SELECT R.id, R.name, R.status, R.vcs_type, R.remote_uri, R.created, R.deleted," +
		" RGM.owner, RGM.name, RGM.hook_id, RGM.hook_secret, RGM.hook_types, RGM.oauth_token" +
		" FROM repositories R LEFT JOIN repository_github_metadatas RGM" +
		" ON R.id=RGM.repository_id WHERE RGM.owner=$1 AND RGM.name=$2"
	row := readHandler.database.QueryRow(query, ownerName, repositoryName)
	return readHandler.scanRepository(row)
}

func (readHandler *ReadHandler) GetAll() ([]resources.Repository, error) {
	query := "SELECT R.id, R.name, R.status, R.vcs_type, R.remote_uri, R.created, R.deleted," +
		" RGM.owner, RGM.name, RGM.hook_id, RGM.hook_secret, RGM.hook_types, RGM.oauth_token" +
		" FROM repositories R LEFT JOIN repository_github_metadatas RGM" +
		" ON R.id=RGM.repository_id WHERE R.id != R.deleted"
	rows, err := readHandler.database.Query(query)
	if err != nil {
		return nil, err
	}

	repositories := make([]resources.Repository, 0, 10)
	for rows.Next() {
		repository, err := readHandler.scanRepository(rows)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, *repository)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return repositories, nil
}
