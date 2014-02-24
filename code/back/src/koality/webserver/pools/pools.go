package pools

import (
	"github.com/gorilla/mux"
	"github.com/mitchellh/goamz/ec2"
	"koality/resources"
	"koality/webserver/middleware"
)

type sanitizedPool struct {
	Id                uint64 `json:"id"`
	Name              string `json:"name"`
	AccessKey         string `json:"accessKey"`
	SecretKey         string `json:"secretKey"`
	Username          string `json:"username"`
	BaseAmiId         string `json:"baseAmiId"`
	SecurityGroupId   string `json:"securityGroupId"`
	VpcSubnetId       string `json:"vpcSubnetId"`
	InstanceType      string `json:"instanceType"`
	NumReadyInstances uint64 `json:"numReadyInstances"`
	NumMaxInstances   uint64 `json:"numMaxInstances"`
	RootDriveSize     uint64 `json:"rootDriveSize"`
	UserData          string `json:"userData"`
}

type sanitizedImage struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type awsSettings struct {
	SecurityGroups []ec2.SecurityGroup `json:"securityGroups"`
	Images         []sanitizedImage    `json:"amis"`
	InstanceTypes  []string            `json:"instanceTypes"`
}

type poolRequestData struct {
	Name                string `json:"name"`
	AccessKey           string `json:"accessKey"`
	SecretKey           string `json:"secretKey"`
	Username            string `json:"username"`
	BaseAmiId           string `json:"baseAmiId"`
	SecurityGroupId     string `json:"securityGroupId"`
	VpcSubnetId         string `json:"vpcSubnetId"`
	InstanceType        string `json:"instanceType"`
	NumReadyInstances   uint64 `json:"numReadyInstances"`
	MaxRunningInstances uint64 `json:"maxRunningInstances"`
	RootDriveSize       uint64 `json:"rootDriveSize"`
	UserData            string `json:"userData"`
}

type PoolsHandler struct {
	resourcesConnection *resources.Connection
}

func New(resourcesConnection *resources.Connection) (*PoolsHandler, error) {
	return &PoolsHandler{resourcesConnection}, nil
}

func (poolsHandler *PoolsHandler) WirePoolsAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/",
		middleware.IsAdminWrapper(poolsHandler.resourcesConnection, poolsHandler.getAll)).
		Methods("GET")
	subrouter.HandleFunc("/{poolId:[0-9]+}",
		middleware.IsAdminWrapper(poolsHandler.resourcesConnection, poolsHandler.get)).
		Methods("GET")
	subrouter.HandleFunc("/getAwsSettings",
		middleware.IsAdminWrapper(poolsHandler.resourcesConnection, poolsHandler.getAwsSettings)).
		Methods("GET")

	subrouter.HandleFunc("/",
		middleware.IsAdminWrapper(poolsHandler.resourcesConnection, poolsHandler.create)).
		Methods("POST")

	subrouter.HandleFunc("/{poolId:[0-9]+}",
		middleware.IsAdminWrapper(poolsHandler.resourcesConnection, poolsHandler.update)).
		Methods("PUT")

	subrouter.HandleFunc("/{poolId:[0-9]+}",
		middleware.IsAdminWrapper(poolsHandler.resourcesConnection, poolsHandler.delete)).
		Methods("DELETE")
}

func (poolsHandler *PoolsHandler) WirePoolsApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{poolId:[0-9]+}", poolsHandler.get).Methods("GET")
	subrouter.HandleFunc("/", poolsHandler.getAll).Methods("GET")
	subrouter.HandleFunc("/{poolId:[0-9]+}", poolsHandler.update).Methods("PUT")
	subrouter.HandleFunc("/", poolsHandler.create).Methods("POST")
	subrouter.HandleFunc("/{poolId:[0-9]+}", poolsHandler.delete).Methods("DELETE")
}

func getSanitizedPool(pool *resources.Ec2Pool) *sanitizedPool {
	return &sanitizedPool{
		Id:                pool.Id,
		Name:              pool.Name,
		AccessKey:         pool.AccessKey,
		SecretKey:         pool.SecretKey,
		Username:          pool.Username,
		BaseAmiId:         pool.BaseAmiId,
		SecurityGroupId:   pool.SecurityGroupId,
		VpcSubnetId:       pool.VpcSubnetId,
		InstanceType:      pool.InstanceType,
		NumReadyInstances: pool.NumReadyInstances,
		NumMaxInstances:   pool.NumMaxInstances,
		RootDriveSize:     pool.RootDriveSize,
		UserData:          pool.UserData,
	}
}

func getSanitizedImage(image ec2.Image) sanitizedImage {
	return sanitizedImage{
		Id:   image.Id,
		Name: image.Name,
	}
}
