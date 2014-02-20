package pools

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (poolsHandler *PoolsHandler) delete(writer http.ResponseWriter, request *http.Request) {
	poolIdString := mux.Vars(request)["poolId"]
	poolId, err := strconv.ParseUint(poolIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse userId: %v", err)
		return
	}

	err = poolsHandler.resourcesConnection.Pools.Delete.DeleteEc2Pool(poolId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}
