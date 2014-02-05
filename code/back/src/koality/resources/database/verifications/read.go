package verifications

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
	subscriptionHandler resources.InternalVerificationsSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalVerificationsSubscriptionHandler) (resources.VerificationsReadHandler, error) {
	return &ReadHandler{database, verifier, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) scanVerification(scannable Scannable) (*resources.Verification, error) {
	verification := new(resources.Verification)

	var mergeTarget, emailToNotify, status, mergeStatus sql.NullString
	err := scannable.Scan(&verification.Id, &verification.RepositoryId, &mergeTarget, &emailToNotify,
		&status, &mergeStatus, &verification.Created, &verification.Started, &verification.Ended,
		&verification.Changeset.Id, &verification.Changeset.RepositoryId, &verification.Changeset.HeadSha,
		&verification.Changeset.BaseSha, &verification.Changeset.HeadMessage, &verification.Changeset.HeadUsername,
		&verification.Changeset.HeadEmail, &verification.Changeset.Created)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchVerificationError{"Unable to find verification"}
	} else if err != nil {
		return nil, err
	}

	if mergeTarget.Valid {
		verification.MergeTarget = mergeTarget.String
	}
	if emailToNotify.Valid {
		verification.EmailToNotify = emailToNotify.String
	}
	if status.Valid {
		verification.Status = status.String
	}
	if mergeStatus.Valid {
		verification.MergeStatus = mergeStatus.String
	}
	return verification, nil
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

func (readHandler *ReadHandler) Get(verificationId uint64) (*resources.Verification, error) {
	query := "SELECT V.id, V.repository_id, V.merge_target, V.email_to_notify," +
		" V.status, V.merge_status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.created" +
		" FROM verifications V JOIN changesets C" +
		" ON V.changeset_id=C.id WHERE V.id=$1"
	row := readHandler.database.QueryRow(query, verificationId)
	return readHandler.scanVerification(row)
}

func (readHandler *ReadHandler) GetTail(repositoryId uint64, offset, results uint32) ([]resources.Verification, error) {
	if err := readHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return nil, err
	} else if err := readHandler.verifier.verifyTailResultsNumber(results); err != nil {
		return nil, err
	}

	query := "SELECT V.id, V.repository_id, V.merge_target, V.email_to_notify," +
		" V.status, V.merge_status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.created" +
		" FROM verifications V JOIN changesets C" +
		" ON V.changeset_id=C.id WHERE V.repository_id=$1" +
		" ORDER BY V.id DESC LIMIT $2 OFFSET $3"
	rows, err := readHandler.database.Query(query, repositoryId, results, offset)
	if err != nil {
		return nil, err
	}

	verifications := make([]resources.Verification, 0, 100)
	for rows.Next() {
		verification, err := readHandler.scanVerification(rows)
		if err != nil {
			return nil, err
		}
		verifications = append(verifications, *verification)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return verifications, nil
}

func (readHandler *ReadHandler) GetChangesetFromShas(headSha, baseSha string) (*resources.Changeset, error) {
	query := "SELECT id, repository_id, head_sha, base_sha, head_message, head_username, head_email, created" +
		" FROM changesets WHERE head_sha=$1 AND base_sha=$2"
	row := readHandler.database.QueryRow(query, headSha, baseSha)
	return readHandler.scanChangeset(row)
}
