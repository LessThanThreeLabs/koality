package verifications

import (
	"database/sql"
	"koality/resources"
)

const (
	initialVerificationStatus = "received"
)

type CreateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewCreateHandler(database *sql.DB, verifier *Verifier) (resources.VerificationsCreateHandler, error) {
	return &CreateHandler{database, verifier}, nil
}

func (createHandler *CreateHandler) Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (uint64, error) {
	if err := createHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return 0, err
	} else if err := createHandler.getChangesetParamsError(headSha, baseSha, headMessage, headUsername, headEmail); err != nil {
		return 0, err
	} else if err := createHandler.getVerificationParamsError(mergeTarget, emailToNotify); err != nil {
		return 0, err
	}

	transaction, err := createHandler.database.Begin()
	if err != nil {
		return 0, err
	}

	changesetId := uint64(0)
	changesetQuery := "INSERT INTO changesets (repository_id, head_sha, base_sha, head_message, head_username, head_email)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err = transaction.QueryRow(changesetQuery, repositoryId, headSha, baseSha, headMessage, headUsername, headEmail).Scan(&changesetId)
	if err != nil {
		transaction.Rollback()
		return 0, err
	}

	verificationId := uint64(0)
	verificationQuery := "INSERT INTO verifications (repository_id, changeset_id, merge_target, email_to_notify, status)" +
		" VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = transaction.QueryRow(verificationQuery, repositoryId, changesetId, mergeTarget, emailToNotify, initialVerificationStatus).Scan(&verificationId)
	if err != nil {
		transaction.Rollback()
		return 0, err
	}

	transaction.Commit()
	return verificationId, nil
}

func (createHandler *CreateHandler) CreateFromChangeset(repositoryId, changesetId uint64, mergeTarget, emailToNotify string) (uint64, error) {
	if err := createHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return 0, err
	} else if err := createHandler.verifier.verifyChangesetExists(changesetId); err != nil {
		return 0, err
	} else if err := createHandler.getVerificationParamsError(mergeTarget, emailToNotify); err != nil {
		return 0, err
	}

	id := uint64(0)
	query := "INSERT INTO verifications (repository_id, changeset_id, merge_target, owner_email, status)" +
		" VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := createHandler.database.QueryRow(query, repositoryId, changesetId, mergeTarget, emailToNotify, initialVerificationStatus).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
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

func (createHandler *CreateHandler) getVerificationParamsError(mergeTarget, emailToNotify string) error {
	if err := createHandler.verifier.verifyMergeTarget(mergeTarget); err != nil {
		return err
	} else if err := createHandler.verifier.verifyEmailToNotify(emailToNotify); err != nil {
		return err
	}
	return nil
}