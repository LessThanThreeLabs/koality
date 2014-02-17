package pools

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (poolsHandler *PoolsHandler) get(writer http.ResponseWriter, request *http.Request) {
	poolIdString := mux.Vars(request)["poolId"]
	poolId, err := strconv.ParseUint(poolIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse poolId: %v", err)
		return
	}

	pool, err := poolsHandler.resourcesConnection.Pools.Read.GetEc2Pool(poolId)
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
