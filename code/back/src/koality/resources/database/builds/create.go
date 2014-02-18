package builds

import (
	"crypto/md5"
	"database/sql"
	"koality/resources"
)

const (
	initialBuildStatus = "declared"
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	readHandler         resources.BuildsReadHandler
	subscriptionHandler resources.InternalBuildsSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.BuildsReadHandler,
	subscriptionHandler resources.InternalBuildsSubscriptionHandler) (resources.BuildsCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, subscriptionHandler}, nil
}

func nilOnZero(id uint64) interface{} {
	if id == 0 {
		return nil
	} else {
		return id
	}
}

func (createHandler *CreateHandler) create(repositoryId, snapshotId, debugInstanceId uint64, headSha, baseSha, headMessage, headUsername, headEmail string, patchContents []byte, emailToNotify, ref string, reuseChangeset, shouldMerge bool) (*resources.Build, error) {
	if err := createHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return nil, err
	} else if err := createHandler.getChangesetParamsError(headSha, baseSha, headMessage, headUsername, headEmail, patchContents); err != nil {
		return nil, err
	} else if err := createHandler.getBuildParamsError(snapshotId, ref, emailToNotify); err != nil {
		return nil, err
	}

	transaction, err := createHandler.database.Begin()
	if err != nil {
		return nil, err
	}

	var ok bool

	changeset, err := createHandler.readHandler.GetChangesetFromShas(headSha, baseSha, patchContents)
	if _, ok = err.(resources.NoSuchChangesetError); err != nil && !ok {
		return nil, err
	} else if !ok && reuseChangeset {
		return createHandler.CreateFromChangeset(repositoryId, changeset.Id, emailToNotify, ref, shouldMerge)
	} else {
		patchHash := md5.Sum(patchContents)

		changesetId := uint64(0)
		changesetQuery := "INSERT INTO changesets (repository_id, head_sha, base_sha, head_message, head_username, head_email, patch_contents, patch_hash)" +
			" VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
		err = transaction.QueryRow(changesetQuery, repositoryId, headSha, baseSha, headMessage, headUsername, headEmail, patchContents, patchHash[:]).Scan(&changesetId)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}

		buildId := uint64(0)
		buildQuery := "INSERT INTO builds (repository_id, snapshot_id, debug_instance_id, changeset_id, ref, should_merge, email_to_notify, status)" +
			" VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"

		err = transaction.QueryRow(buildQuery, repositoryId, nilOnZero(snapshotId), nilOnZero(debugInstanceId), changesetId, ref, shouldMerge, emailToNotify, initialBuildStatus).Scan(&buildId)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}

		transaction.Commit()

		build, err := createHandler.readHandler.Get(buildId)
		if err != nil {
			return nil, err
		}

		return build, nil
	}

	return nil, nil
}

func (createHandler *CreateHandler) Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail string, patchContents []byte, emailToNotify, ref string, reuseChangeset, shouldMerge bool) (*resources.Build, error) {
	build, err := createHandler.create(repositoryId, 0, 0, headSha, baseSha, headMessage, headUsername, headEmail, patchContents, emailToNotify, ref, reuseChangeset, shouldMerge)
	if err == nil {
		createHandler.subscriptionHandler.FireCreatedEvent(build)
	}

	return build, err
}

func (createHandler *CreateHandler) CreateForSnapshot(repositoryId, snapshotId uint64, headSha, baseSha, headMessage, headUsername, headEmail, emailToNotify, ref string) (*resources.Build, error) {
	return createHandler.create(repositoryId, snapshotId, 0, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, true, false)
}

func (createHandler *CreateHandler) CreateForDebugInstance(repositoryId, debugInstanceId uint64, headSha, baseSha, headMessage, headUsername, headEmail string, emailToNotify, ref string) (*resources.Build, error) {
	return createHandler.create(repositoryId, 0, debugInstanceId, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, true, false)
}

func (createHandler *CreateHandler) CreateFromChangeset(repositoryId, changesetId uint64, emailToNotify, ref string, shouldMerge bool) (*resources.Build, error) {
	if err := createHandler.verifier.verifyRepositoryExists(repositoryId); err != nil {
		return nil, err
	} else if err := createHandler.verifier.verifyChangesetExists(changesetId); err != nil {
		return nil, err
	} else if err := createHandler.getBuildParamsError(0, ref, emailToNotify); err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO builds (repository_id, snapshot_id, changeset_id, ref, should_merge, email_to_notify, status)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"

	err := createHandler.database.QueryRow(query, repositoryId, nil, changesetId, ref, shouldMerge, emailToNotify, initialBuildStatus).Scan(&id)
	if err != nil {
		return nil, err
	}

	build, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(build)
	return build, nil
}

func (createHandler *CreateHandler) getChangesetParamsError(headSha, baseSha, headMessage, headUsername, headEmail string, patchContents []byte) error {
	if err := createHandler.verifier.verifyHeadSha(headSha); err != nil {
		return err
	} else if err := createHandler.verifier.verifyBaseSha(baseSha); err != nil {
		return err
	} else if err := createHandler.verifier.verifyHeadMessage(headMessage); err != nil {
		return err
	} else if err := createHandler.verifier.verifyHeadUsername(headUsername); err != nil {
		return err
	} else if err := createHandler.verifier.verifyHeadEmail(headEmail); err != nil {
		return err
	} else if err := createHandler.verifier.verifyPatchContents(patchContents); err != nil {
		return err
	}
	return nil
}

func (createHandler *CreateHandler) getBuildParamsError(snapshotId uint64, ref, emailToNotify string) error {
	if err := createHandler.verifier.verifyRef(ref); err != nil {
		return err
	} else if err := createHandler.verifier.verifyEmailToNotify(emailToNotify); err != nil {
		return err
	} else if err := createHandler.verifier.verifySnapshotExists(snapshotId); snapshotId != 0 && err != nil {
		return err
	}
	return nil
}
