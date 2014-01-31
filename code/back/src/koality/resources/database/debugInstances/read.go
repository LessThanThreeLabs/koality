package debugInstances

import (
	"database/sql"
	"koality/resources"
)

type Scannable interface {
	Scan(dest ...interface{}) error
}

type ReadHandler struct {
	database            *sql.DB
	subscriptionHandler resources.InternalDebugInstancesSubscriptionHandler
}

func NewReadHandler(database *sql.DB, subscriptionHandler resources.InternalDebugInstancesSubscriptionHandler) (resources.DebugInstancesReadHandler, error) {
	return &ReadHandler{database, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) scanDebugInstance(scannable Scannable) (*resources.DebugInstance, error) {
	debugInstance := new(resources.DebugInstance)

	var deletedId uint64
	err := scannable.Scan(&debugInstance.Id, &debugInstance.PoolId, &debugInstance.InstanceId, &debugInstance.Expires, &deletedId, &debugInstance.VerificationId)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchDebugInstanceError{"Unable to find debug instance"}
	} else if err != nil {
		return nil, err
	}

	debugInstance.IsDeleted = debugInstance.Id == deletedId
	return debugInstance, nil
}

func (readHandler *ReadHandler) Get(debugInstanceId uint64) (*resources.DebugInstance, error) {
	query := "SELECT D.id, D.pool_id, D.instance_id, D.expires, D.deleted, V.id" +
		" FROM debug_instances D JOIN verifications V" +
		" ON V.debug_instance_id=D.id WHERE D.id=$1"
	row := readHandler.database.QueryRow(query, debugInstanceId)
	return readHandler.scanDebugInstance(row)
}

func (readHandler *ReadHandler) GetAllRunning() ([]resources.DebugInstance, error) {
	query := "SELECT id, pool_id, instance_id, expires, deleted FROM debug_instances" +
		" WHERE deleted = $1"
	rows, err := readHandler.database.Query(query, 0)
	if err != nil {
		return nil, err
	}

	debugInstances := make([]resources.DebugInstance, 0, 10)
	for rows.Next() {
		debugInstance, err := readHandler.scanDebugInstance(rows)
		if err != nil {
			return nil, err
		}
		debugInstances = append(debugInstances, *debugInstance)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return debugInstances, nil
}
