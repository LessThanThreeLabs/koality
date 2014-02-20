package main

import (
	"koality/internalapi"
	"log"
	"net/rpc"
	"os"
	"strconv"
)

func main() {
	userId, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	// Should we be using this?
	userId = userId

	repositoryName := os.Args[2]
	ref := os.Args[3]
	baseSha := os.Args[4]
	headSha := os.Args[5]

	var buildId uint64

	client, err := rpc.Dial("unix", internalapi.RpcSocket)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Call("BuildCreator.CreateBuild", internalapi.CreateBuildArg{repositoryName, ref, baseSha, headSha}, &buildId)
	if err != nil {
		log.Fatal(err)
	}
}
