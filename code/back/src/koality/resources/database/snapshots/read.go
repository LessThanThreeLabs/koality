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

	var imageId sql.NullString
	var deletedId uint64
	err := scannable.Scan(&snapshot.Id, &snapshot.PoolId, &imageId, &snapshot.ImageType, &snapshot.Status,
		&snapshot.Created, &snapshot.Started, &snapshot.Ended, &deletedId)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchSnapshotError{"Unable to find snapshot"}
	} else if err != nil {
		return nil, err
	}

	snapshot.IsDeleted = snapshot.Id == deletedId

	if imageId.Valid {
		snapshot.ImageId = imageId.String
	}
	return snapshot, nil
}

func (readHandler *ReadHandler) Get(snapshotId uint64) (*resources.Snapshot, error) {
	query := "SELECT id, pool_id, image_id, image_type, status, created, started, ended, deleted" +
		" FROM snapshots WHERE id=$1"
	row := readHandler.database.QueryRow(query, snapshotId)
	return readHandler.scanSnapshot(row)
}

func (readHandler *ReadHandler) GetByImageId(imageId string) (*resources.Snapshot, error) {
	query := "SELECT id, pool_id, image_id, image_type, status, created, started, ended, deleted" +
		" FROM snapshots WHERE image_id=$1"
	row := readHandler.database.QueryRow(query, imageId)
	return readHandler.scanSnapshot(row)
}

func (readHandler *ReadHandler) GetAllForPool(poolId uint64) ([]resources.Snapshot, error) {
	query := "SELECT id, pool_id, image_id, image_type, status, created, started, ended, deleted" +
		" FROM snapshots WHERE pool_id = $1" +
		" ORDER BY id DESC"
	rows, err := readHandler.database.Query(query, poolId)
	if err != nil {
		return nil, err
	}

	Snapshots := make([]resources.Snapshot, 0, 10)
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
