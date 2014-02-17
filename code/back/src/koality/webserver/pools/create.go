package pools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (poolsHandler *PoolsHandler) create(writer http.ResponseWriter, request *http.Request) {
	name := request.PostFormValue("name")
	accessKey := request.PostFormValue("accessKey")
	secretKey := request.PostFormValue("secretkey")
	username := request.PostFormValue("username")
	baseAmiId := request.PostFormValue("baseAmiId")
	securityGroupId := request.PostFormValue("securityGroupId")
	vpcSubnetId := request.PostFormValue("vpcSubnetId")
	instanceType := request.PostFormValue("instanceType")
	numReadyInstancesString := request.PostFormValue("numReadyInstances")
	numReadyInstances, err := strconv.ParseUint(numReadyInstancesString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse numReadyInstances: %v", err)
		return
	}

	numMaxInstancesString := request.PostFormValue("numMaxInstances")
	numMaxInstances, err := strconv.ParseUint(numMaxInstancesString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse numMaxInstances: %v", err)
		return
	}

	rootDriveSizeString := request.PostFormValue("rootDriveSize")
	rootDriveSize, err := strconv.ParseUint(rootDriveSizeString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse rootDriveSize: %v", err)
		return
	}

	userData := request.PostFormValue("userData")

	pool, err := poolsHandler.resourcesConnection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)

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
