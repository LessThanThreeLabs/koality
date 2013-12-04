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
	database *sql.DB
}

func NewReadHandler(database *sql.DB) (resources.RepositoriesReadHandler, error) {
	return &ReadHandler{database}, nil
}

func (readHandler *ReadHandler) scanRepository(scannable Scannable) (*resources.Repository, error) {
	repository := new(resources.Repository)

	var gitHubHookId sql.NullInt64
	var gitHubOwner, gitHubName, gitHubHookSecret, gitHubHookTypes sql.NullString
	err := scannable.Scan(&repository.Id, &repository.Name, &repository.VcsType,
		&repository.LocalUri, &repository.RemoteUri, &repository.Created,
		&gitHubOwner, &gitHubName, &gitHubHookId, &gitHubHookSecret, &gitHubHookTypes)
	if err == sql.ErrNoRows {
		errorText := "Unable to find repository"
		return nil, resources.NoSuchRepositoryError{errorText}
	} else if err != nil {
		return nil, err
	}

	if gitHubOwner.Valid && gitHubName.Valid {
		repository.GitHub = new(resources.GitHubMetadata)
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

	return repository, nil
}

func (readHandler *ReadHandler) Get(repositoryId uint64) (*resources.Repository, error) {
	query := "SELECT R.id, R.name, R.vcs_type, R.local_uri, R.remote_uri, R.created," +
		" RGM.owner, RGM.name, RGM.hook_id, RGM.hook_secret, RGM.hook_types" +
		" FROM repositories R LEFT JOIN repository_github_metadatas RGM" +
		" ON R.id=RGM.repository_id WHERE R.id=$1"
	row := readHandler.database.QueryRow(query, repositoryId)
	return readHandler.scanRepository(row)
}

func (readHandler *ReadHandler) GetAll() ([]resources.Repository, error) {
	query := "SELECT R.id, R.name, R.vcs_type, R.local_uri, R.remote_uri, R.created," +
		" RGM.owner, RGM.name, RGM.hook_id, RGM.hook_secret, RGM.hook_types" +
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
