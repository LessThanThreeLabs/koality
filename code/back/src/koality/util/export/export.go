package export

import (
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"io/ioutil"
	"koality/resources"
	"koality/util/find"
	"os"
	"path/filepath"
	"time"
)

type ExportOutput struct {
	Exports []resources.Export
	Error   error
}

func Export(accessKey, secretKey, bucketName, exportPrefix string, region aws.Region, globPatterns []string) ([]resources.Export, error) {
	token := ""
	expiration := time.Time{}
	auth, err := aws.GetAuth(accessKey, secretKey, token, expiration)
	if err != nil {
		return nil, err
	}

	var exports []resources.Export
	s3Obj := s3.New(auth, region)
	bucket := s3Obj.Bucket(bucketName)
	if err = bucket.PutBucket(s3.PublicRead); err != nil {
		s3Err, ok := err.(*s3.Error)
		if !(ok && s3Err.Code == "BucketAlreadyOwnedByYou") {
		// return error unless the error was that the bucket is already owned by us
			return nil, err
		}
	}

	uploadFile := func(path string, fileInfo os.FileInfo, err error) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		s3Path := exportPrefix + path
		contentType := ""
		err = bucket.Put(s3Path, data, contentType, s3.PublicRead, s3.Options{})
		if err != nil {
			return err
		}

		exports = append(exports, resources.Export{
			BucketName: bucketName,
			Path:       path,
			Key:        s3Path,
		})
		return nil
	}

	for _, globPattern := range globPatterns {
		paths, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, err
		}

		for _, path := range paths {
			fileInfo, err := os.Stat(path)
			if err != nil {
				return nil, err
			}

			if fileInfo.Mode().IsRegular() {
				err = uploadFile(path, fileInfo, nil)
			} else {
				err = find.Find(path, "*", uploadFile)
			}
			if err != nil {
				return nil, err
			}
		}
	}

	return exports, nil
}
