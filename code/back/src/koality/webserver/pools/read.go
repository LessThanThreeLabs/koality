package pools

import (
	"encoding/json"
	"fmt"
	"github.com/LessThanThreeLabs/goamz/aws"
	"github.com/LessThanThreeLabs/goamz/ec2"
	"github.com/gorilla/mux"
	"koality/resources/database/pools"
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

func (poolsHandler *PoolsHandler) getAll(writer http.ResponseWriter, request *http.Request) {
	pools, err := poolsHandler.resourcesConnection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedPools := make([]*sanitizedPool, len(pools))
	for i, pool := range pools {
		sanitizedPools[i] = getSanitizedPool(&pool)
	}
	jsonedPools, err := json.Marshal(sanitizedPools)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, "Unable to stringify:", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedPools)
}

func (poolsHandler *PoolsHandler) getAwsSettings(writer http.ResponseWriter, request *http.Request) {
	auth := aws.Auth{
		AccessKey: request.FormValue("accessKey"),
		SecretKey: request.FormValue("secretKey"),
	}
	region := aws.USWest2 // TODO (bbland): change to USEast
	ec2Conn := ec2.New(auth, region)
	securityGroupsResp, err := ec2Conn.SecurityGroups(nil, nil)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	imagesResp, err := ec2Conn.Images(nil, []string{"self"}, nil)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	var securityGroups []ec2.SecurityGroup
	for _, securityGroupInfo := range securityGroupsResp.Groups {
		securityGroups = append(securityGroups, securityGroupInfo.SecurityGroup)
	}

	var sanitizedImages []sanitizedImage
	for _, image := range imagesResp.Images {
		sanitizedImages = append(sanitizedImages, getSanitizedImage(image))
	}
	awsSettings := awsSettings{
		SecurityGroups: securityGroups,
		Images:         sanitizedImages,
		InstanceTypes:  pools.AllowedEc2InstanceTypes,
	}
	jsonedAwsSettings, err := json.Marshal(awsSettings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedAwsSettings)
}
