package pools

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (poolsHandler *PoolsHandler) create(writer http.ResponseWriter, request *http.Request) {
	poolData := new(poolRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(poolData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	pool, err := poolsHandler.resourcesConnection.Pools.Create.CreateEc2Pool(poolData.Name, poolData.AccessKey, poolData.SecretKey, poolData.Username, poolData.BaseAmiId, poolData.SecurityGroupId, poolData.VpcSubnetId, poolData.InstanceType, poolData.NumReadyInstances, poolData.MaxRunningInstances, poolData.RootDriveSize, poolData.UserData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedPool := getSanitizedPool(pool)
	jsonedPool, err := json.Marshal(sanitizedPool)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedPool)
}
