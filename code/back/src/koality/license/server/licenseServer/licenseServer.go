package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"io/ioutil"
	"koality/license/server"
	"koality/util/pathtranslator"
	"path"
)

const (
	port = 9000
)

func main() {
	// TODO (bbland): figure out what path to use here
	configPath, err := pathtranslator.TranslatePathAndCheckExists(path.Join(
		"code", "back", "src", "koality", "license", "server", "licenseServer", "license-server-config.json"))
	if err != nil {
		panic(err)
	}
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	config := new(licenseserver.ServerConfiguration)
	if err = json.Unmarshal(configBytes, config); err != nil {
		panic(err)
	}

	database, err := licenseserver.GetDatabaseConnection(config.Database)
	if err != nil {
		panic(err)
	}

	awsAuth, err := aws.GetAuth(config.S3.AwsAccessKeyId, config.S3.AwsSecretAccessKey)
	if err != nil {
		panic(err)
	}

	s3Connection := s3.New(awsAuth, aws.USEast)
	bucket := s3Connection.Bucket(config.S3.BucketName)
	if err = bucket.PutBucket(s3.Private); err != nil {
		s3Err, ok := err.(*s3.Error)
		if !(ok && s3Err.Code == "BucketAlreadyOwnedByYou") {
			panic(err)
		}
	}

	licenseServer := licenseserver.New(database, bucket, port)

	fmt.Printf("Starting license server on port %d...\n", port)

	if err = licenseServer.Start(); err != nil {
		panic(err)
	}
}
