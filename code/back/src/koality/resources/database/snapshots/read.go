package snapshots

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
	subscriptionHandler resources.InternalSnapshotsSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalSnapshotsSubscriptionHandler) (resources.SnapshotsReadHandler, error) {
	return &ReadHandler{database, verifier, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) scanSnapshot(scannable Scannable) (*resources.Snapshot, error) {
	snapshot := new(resources.Snapshot)
	err := scannable.Scan(&snapshot.PoolId, &snapshot.Id, &snapshot.ImageId, &snapshot.ImageType, &snapshot.Status, &snapshot.Deleted,
		&snapshot.Created, &snapshot.Started, &snapshot.Ended)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchUserError{"Unable to find snapshot"}
	} else if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (readHandler *ReadHandler) GetSnapshot(snapshotId uint64) (*resources.Snapshot, error) {
	query := "SELECT id, poolId, imageId, imageType, status, deleted, created, started, ended" +
		" FROM snapshots WHERE id=$1"
	row := readHandler.database.QueryRow(query, snapshotId)
	return readHandler.scanSnapshot(row)
}

func (readHandler *ReadHandler) GetSnapshotsForPoolId(poolId uint64) ([]resources.Snapshot, error) {
	query := "SELECT id, poolId, imageId, imageType, status, deleted, created, started, ended" +
		" FROM snapshots WHERE pool_id = $1" +
		" ORDER BY id DESC"
	rows, err := readHandler.database.Query(query, poolId)
	if err != nil {
		return nil, err
	}

	Snapshots := make([]resources.Snapshot, 0, 1)
	for rows.Next() {
		snapshot, err := readHandler.scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		Snapshots = append(Snapshots, *snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return Snapshots, nil
}
