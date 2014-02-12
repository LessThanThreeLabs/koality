package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"koality/util/export"
	"os"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	accessKey := os.Args[1]
	secretKey := os.Args[2]
	bucketName := os.Args[3]
	exportPrefix := os.Args[4]
	regionStr := os.Args[5]
	globPatterns := os.Args[6:]

	region, ok := aws.Regions[regionStr]
	if !ok {
		err = errors.New("region is invalid")
		return
	}

	exports, err := export.Export(accessKey, secretKey, bucketName, exportPrefix, region, globPatterns)
	if err != nil {
		return
	}

	exportBytes, err := json.Marshal(export.ExportOutput{Exports: exports, Error: err})
	if err != nil {
		return
	}

	os.Stdout.Write(exportBytes)
}
