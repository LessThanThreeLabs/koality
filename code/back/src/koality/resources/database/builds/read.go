package builds

import (
	"database/sql"
	"koality/resources"
)

type Scannable interface {
	Scan(dest ...interface{}) error
}

type ReadHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalBuildsSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalBuildsSubscriptionHandler) (resources.BuildsReadHandler, error) {
	return &ReadHandler{database, verifier, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) scanBuild(scannable Scannable) (*resources.Build, error) {
	build := new(resources.Build)

	var mergeTarget, emailToNotify, status, mergeStatus sql.NullString
	err := scannable.Scan(&build.Id, &build.RepositoryId, &mergeTarget, &emailToNotify,
		&status, &mergeStatus, &build.Created, &build.Started, &build.Ended,
		&build.Changeset.Id, &build.Changeset.RepositoryId, &build.Changeset.HeadSha,
		&build.Changeset.BaseSha, &build.Changeset.HeadMessage, &build.Changeset.HeadUsername,
		&build.Changeset.HeadEmail, &build.Changeset.Created)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchBuildError{"Unable to find build"}
	} else if err != nil {
		return nil, err
	}

	if mergeTarget.Valid {
		build.MergeTarget = mergeTarget.String
	}
	if emailToNotify.Valid {
		build.EmailToNotify = emailToNotify.String
	}
	if status.Valid {
		build.Status = status.String
	}
	if mergeStatus.Valid {
		build.MergeStatus = mergeStatus.String
	}
	return build, nil
}

func (readHandler *ReadHandler) scanChangeset(scannable Scannable) (*resources.Changeset, error) {
	changeset := new(resources.Changeset)

	err := scannable.Scan(&changeset.Id, &changeset.RepositoryId, &changeset.HeadSha,
		&changeset.BaseSha, &changeset.HeadMessage, &changeset.HeadUsername,
		&changeset.HeadEmail, &changeset.Created)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchChangesetError{"Unable to find changeset"}
	} else if err != nil {
		return nil, err
	}
	return changeset, nil
}

func (readHandler *ReadHandler) Get(buildId uint64) (*resources.Build, error) {
	query := "SELECT V.id, V.repository_id, V.merge_target, V.email_to_notify," +
		" V.status, V.merge_status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.created" +
		" FROM builds V JOIN changesets C" +
		" ON V.changeset_id=C.id WHERE V.id=$1"
	row := readHandler.database.QueryRow(query, buildId)
	return readHandler.scanBuild(row)
}

func (readHandler *ReadHandler) GetTail(repositoryId uint64, offset, results uint32) ([]resources.Build, error) {
	if err := readHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return nil, err
	} else if err := readHandler.verifier.verifyTailResultsNumber(results); err != nil {
		return nil, err
	}

	query := "SELECT V.id, V.repository_id, V.merge_target, V.email_to_notify," +
		" V.status, V.merge_status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.created" +
		" FROM builds V JOIN changesets C" +
		" ON V.changeset_id=C.id WHERE V.repository_id=$1" +
		" ORDER BY V.id DESC LIMIT $2 OFFSET $3"
	rows, err := readHandler.database.Query(query, repositoryId, results, offset)
	if err != nil {
		return nil, err
	}

	builds := make([]resources.Build, 0, 100)
	for rows.Next() {
		build, err := readHandler.scanBuild(rows)
		if err != nil {
			return nil, err
		}
		builds = append(builds, *build)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return builds, nil
}

func (readHandler *ReadHandler) GetChangesetFromShas(headSha, baseSha string) (*resources.Changeset, error) {
	query := "SELECT id, repository_id, head_sha, base_sha, head_message, head_username, head_email, created" +
		" FROM changesets WHERE head_sha=$1 AND base_sha=$2"
	row := readHandler.database.QueryRow(query, headSha, baseSha)
	return readHandler.scanChangeset(row)
}
