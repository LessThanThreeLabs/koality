package builds

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"koality/resources"
	"strings"
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

	var ref, emailToNotify, status sql.NullString
	var snapshotId, debugInstanceId sql.NullInt64
	err := scannable.Scan(&build.Id, &build.RepositoryId, &snapshotId, &debugInstanceId,
		&ref, &build.ShouldMerge, &emailToNotify, &status, &build.Created, &build.Started, &build.Ended,
		&build.Changeset.Id, &build.Changeset.RepositoryId, &build.Changeset.HeadSha,
		&build.Changeset.BaseSha, &build.Changeset.HeadMessage, &build.Changeset.HeadUsername,
		&build.Changeset.HeadEmail, &build.Changeset.PatchContents, &build.Changeset.Created)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchBuildError{"Unable to find build"}
	} else if err != nil {
		return nil, err
	}

	if ref.Valid {
		build.Ref = ref.String
	}
	if emailToNotify.Valid {
		build.EmailToNotify = emailToNotify.String
	}
	if status.Valid {
		build.Status = status.String
	}
	if snapshotId.Valid {
		build.SnapshotId = uint64(snapshotId.Int64)
	}
	if debugInstanceId.Valid {
		build.SnapshotId = uint64(debugInstanceId.Int64)
	}

	return build, nil
}

func (readHandler *ReadHandler) scanChangeset(scannable Scannable) (*resources.Changeset, error) {
	changeset := new(resources.Changeset)

	err := scannable.Scan(&changeset.Id, &changeset.RepositoryId, &changeset.HeadSha,
		&changeset.BaseSha, &changeset.HeadMessage, &changeset.HeadUsername,
		&changeset.HeadEmail, &changeset.PatchContents, &changeset.Created)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchChangesetError{"Unable to find changeset"}
	} else if err != nil {
		return nil, err
	}
	return changeset, nil
}

func (readHandler *ReadHandler) Get(buildId uint64) (*resources.Build, error) {
	query := "SELECT V.id, V.repository_id, V.snapshot_id, V.debug_instance_id, V.ref," +
		" V.should_merge, V.email_to_notify, V.status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.patch_contents, C.created" +
		" FROM builds V JOIN changesets C" +
		" ON V.changeset_id=C.id WHERE V.id=$1"
	row := readHandler.database.QueryRow(query, buildId)
	return readHandler.scanBuild(row)
}

func (readHandler *ReadHandler) GetForSnapshot(snapshotId uint64) ([]resources.Build, error) {
	if err := readHandler.verifier.verifySnapshotExists(snapshotId); err != nil {
		return nil, err
	}

	query := "SELECT V.id, V.repository_id, V.snapshot_id, V.debug_instance_id, V.ref," +
		" V.should_merge, V.email_to_notify, V.status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.patch_contents, C.created" +
		" FROM builds V JOIN changesets C" +
		" ON V.changeset_id=C.id WHERE V.snapshot_id=$1"
	rows, err := readHandler.database.Query(query, snapshotId)
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

func (readHandler *ReadHandler) GetBuilds(repositoryIds []uint64, searchParams string, all bool, numResults, offset, userId uint64) ([]resources.Build, error) {
	repositoryIdStrings := []string{}
	for _, repositoryId := range repositoryIds {
		repositoryIdStrings = append(repositoryIdStrings, fmt.Sprintf("%d", repositoryId))
	}
	repositoryIdsString := strings.Join(repositoryIdStrings, ",")
	searchCondition := fmt.Sprintf("(LOWER(C.head_sha) LIKE LOWER('%s%%') OR LOWER(B.ref)=LOWER($1))", searchParams)
	if all {
		searchCondition = fmt.Sprintf("%s OR LOWER($1)=LOWER(U.first_name) OR LOWER($1)=LOWER(U.last_name)", searchCondition)
	} else {
		searchCondition = fmt.Sprintf("%s AND U.id=%d", searchCondition, userId)
	}

	query := fmt.Sprintf("SELECT B.id, B.repository_id, B.ref, B.should_merge,"+
		" B.email_to_notify, B.status, B.created, B.started, B.ended,"+
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username,"+
		" C.head_email, C.patch_contents, C.created"+
		" FROM builds B"+
		" JOIN changesets C ON B.changeset_id=C.id"+
		" JOIN users U ON B.email_to_notify=U.email"+
		" WHERE B.repository_id IN (%s) AND (%s)"+
		" ORDER BY B.id DESC LIMIT $2 OFFSET $3",
		repositoryIdsString, searchCondition)
	rows, err := readHandler.database.Query(query, searchParams, numResults, offset)
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

func (readHandler *ReadHandler) GetTail(repositoryId uint64, offset, results uint32) ([]resources.Build, error) {
	if err := readHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return nil, err
	} else if err := readHandler.verifier.verifyTailResultsNumber(results); err != nil {
		return nil, err
	}

	query := "SELECT V.id, V.repository_id, V.snapshot_id, V.debug_instance_id, V.ref," +
		" V.should_merge, V.email_to_notify, V.status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.patch_contents, C.created" +
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

func (readHandler *ReadHandler) GetChangesetFromShas(headSha, baseSha string, patchContents []byte) (*resources.Changeset, error) {
	patchHash := md5.New().Sum(patchContents)
	query := "SELECT id, repository_id, head_sha, base_sha, head_message, head_username, head_email, patch_contents, created" +
		" FROM changesets WHERE head_sha=$1 AND base_sha=$2 AND patch_hash=$3"
	row := readHandler.database.QueryRow(query, headSha, baseSha, patchHash)
	return readHandler.scanChangeset(row)
}
