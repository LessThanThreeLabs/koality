package pools

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (poolsHandler *PoolsHandler) update(writer http.ResponseWriter, request *http.Request) {
	poolIdString := mux.Vars(request)["poolId"]
	poolId, err := strconv.ParseUint(poolIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse poolId: %v", err)
	}

	poolData := new(poolRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(poolData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = poolsHandler.resourcesConnection.Pools.Update.SetEc2Settings(poolId, poolData.AccessKey, poolData.SecretKey, poolData.Username, poolData.BaseAmiId, poolData.SecurityGroupId, poolData.VpcSubnetId, poolData.InstanceType, poolData.NumReadyInstances, poolData.MaxRunningInstances, poolData.RootDriveSize, poolData.UserData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}
