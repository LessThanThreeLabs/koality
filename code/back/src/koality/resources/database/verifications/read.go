package verifications

import (
	"database/sql"
	"errors"
	"koality/resources"
)

type ReadHandler struct {
	database *sql.DB
}

func NewReadHandler(database *sql.DB) (resources.VerificationsReadHandler, error) {
	return &ReadHandler{database}, nil
}

func (readHandler *ReadHandler) Get(verificationId uint64) (*resources.Verification, error) {
	verification := new(resources.Verification)
	var mergeTarget, emailToNotify, verificationStatus, mergeStatus sql.NullString

	query := "SELECT V.id, V.repository_id, V.merge_target, V.email_to_notify," +
		" V.status, V.merge_status, V.created, V.started, V.ended," +
		" C.id, C.repository_id, C.head_sha, C.base_sha, C.head_message, C.head_username," +
		" C.head_email, C.created" +
		" FROM verifications V JOIN changesets C" +
		" ON V.changeset_id=C.id WHERE V.id=$1"
	row := readHandler.database.QueryRow(query, verificationId)
	err := row.Scan(&verification.Id, &verification.RepositoryId, &mergeStatus, &emailToNotify,
		&verificationStatus, &mergeStatus, &verification.Created, &verification.Started, &verification.Ended,
		&verification.Changeset.Id, &verification.Changeset.RepositoryId, &verification.Changeset.HeadSha,
		&verification.Changeset.BaseSha, &verification.Changeset.HeadMessage, &verification.Changeset.HeadUsername,
		&verification.Changeset.HeadEmail, &verification.Changeset.Created)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchVerificationError{errors.New("Unable to find verification")}
	} else if err != nil {
		return nil, err
	}

	if mergeTarget.Valid {
		verification.MergeTarget = mergeTarget.String
	}
	if emailToNotify.Valid {
		verification.EmailToNotify = emailToNotify.String
	}
	if verificationStatus.Valid {
		verification.VerificationStatus = verificationStatus.String
	}
	if mergeStatus.Valid {
		verification.MergeStatus = mergeStatus.String
	}

	return verification, nil
}
