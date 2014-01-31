package verifications

import (
	"database/sql"
	"koality/resources"
)

const (
	initialVerificationStatus = "declared"
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	readHandler         resources.VerificationsReadHandler
	subscriptionHandler resources.InternalVerificationsSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.VerificationsReadHandler,
	subscriptionHandler resources.InternalVerificationsSubscriptionHandler) (resources.VerificationsCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, subscriptionHandler}, nil
}

func nilOnZero(id uint64) interface{} {
	if id == 0 {
		return nil
	} else {
		return id
	}
}

func (createHandler *CreateHandler) create(repositoryId, snapshotId, debugInstanceId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (*resources.Verification, error) {
	if err := createHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return nil, err
	} else if err := createHandler.getChangesetParamsError(headSha, baseSha, headMessage, headUsername, headEmail); err != nil {
		return nil, err
	} else if err := createHandler.getVerificationParamsError(snapshotId, mergeTarget, emailToNotify); err != nil {
		return nil, err
	}

	transaction, err := createHandler.database.Begin()
	if err != nil {
		return nil, err
	}

	changesetId := uint64(0)
	changesetQuery := "INSERT INTO changesets (repository_id, head_sha, base_sha, head_message, head_username, head_email)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err = transaction.QueryRow(changesetQuery, repositoryId, headSha, baseSha, headMessage, headUsername, headEmail).Scan(&changesetId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	verificationId := uint64(0)
	verificationQuery := "INSERT INTO verifications (repository_id, snapshot_id, debug_instance_id, changeset_id, merge_target, email_to_notify, status)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"

	err = transaction.QueryRow(verificationQuery, repositoryId, nilOnZero(snapshotId), nilOnZero(debugInstanceId), changesetId, mergeTarget, emailToNotify, initialVerificationStatus).Scan(&verificationId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	transaction.Commit()

	verification, err := createHandler.readHandler.Get(verificationId)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(verification)
	return verification, nil
}

func (createHandler *CreateHandler) Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (*resources.Verification, error) {
	return createHandler.create(repositoryId, 0, 0, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
}

func (createHandler *CreateHandler) CreateForSnapshot(repositoryId, snapshotId uint64, headSha, baseSha, headMessage, headUsername, headEmail, emailToNotify string) (*resources.Verification, error) {
	return createHandler.create(repositoryId, snapshotId, 0, headSha, baseSha, headMessage, headUsername, headEmail, "", emailToNotify)
}

func (createHandler *CreateHandler) CreateForDebugInstance(repositoryId, debugInstanceId uint64, headSha, baseSha, headMessage, headUsername, headEmail, emailToNotify string) (*resources.Verification, error) {
	return createHandler.create(repositoryId, 0, debugInstanceId, headSha, baseSha, headMessage, headUsername, headEmail, "", emailToNotify)
}

func (createHandler *CreateHandler) CreateFromChangeset(repositoryId, changesetId uint64, mergeTarget, emailToNotify string) (*resources.Verification, error) {
	if err := createHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return nil, err
	} else if err := createHandler.verifier.verifyChangesetExists(changesetId); err != nil {
		return nil, err
	} else if err := createHandler.getVerificationParamsError(0, mergeTarget, emailToNotify); err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO verifications (repository_id, snapshot_id, changeset_id, merge_target, email_to_notify, status)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"

	err := createHandler.database.QueryRow(query, repositoryId, nil, changesetId, mergeTarget, emailToNotify, initialVerificationStatus).Scan(&id)
	if err != nil {
		return nil, err
	}

	verification, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(verification)
	return verification, nil
}

func (createHandler *CreateHandler) getChangesetParamsError(headSha, baseSha, headMessage, headUsername, headEmail string) error {
	if err := createHandler.verifier.verifyHeadSha(headSha); err != nil {
		return err
	} else if err := createHandler.verifier.verifyBaseSha(baseSha); err != nil {
		return err
	} else if err := createHandler.verifier.verifyShaPairDoesNotExist(headSha, baseSha); err != nil {
		return err
	} else if err := createHandler.verifier.verifyHeadMessage(headMessage); err != nil {
		return err
	} else if err := createHandler.verifier.verifyHeadUsername(headUsername); err != nil {
		return err
	} else if err := createHandler.verifier.verifyHeadEmail(headEmail); err != nil {
		return err
	}
	return nil
}

func (createHandler *CreateHandler) getVerificationParamsError(snapshotId uint64, mergeTarget, emailToNotify string) error {
	if err := createHandler.verifier.verifyMergeTarget(mergeTarget); err != nil {
		return err
	} else if err := createHandler.verifier.verifyEmailToNotify(emailToNotify); err != nil {
		return err
	} else if err := createHandler.verifier.verifySnapshotExists(snapshotId); snapshotId != 0 && err != nil {
		return err
	}
	return nil
}
