package settings

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
)

const (
	awsAccessKeyLength    = 20
	awsSecretKeyLength    = 40
	minS3BucketNameLength = 3
	s3BucketNameRegex     = "^[-_a-zA-Z0-9\\.]+$"
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyAwsAccessKey(accessKey string) error {
	if len(accessKey) != awsAccessKeyLength {
		return fmt.Errorf("Access key must be %d characters long", awsAccessKeyLength)
	}
	return nil
}

func (verifier *Verifier) verifyAwsSecretKey(secretKey string) error {
	if len(secretKey) != awsSecretKeyLength {
		return fmt.Errorf("Secret key must be %d characters long", awsSecretKeyLength)
	}
	return nil
}

func (verifier *Verifier) verifyS3BucketName(bucketName string) error {
	if len(bucketName) < minS3BucketNameLength {
		return fmt.Errorf("Bucket name must be %d characters long", minS3BucketNameLength)
	} else if ok, err := regexp.MatchString(s3BucketNameRegex, bucketName); !ok || err != nil {
		return errors.New("Bucket name must match regex: " + s3BucketNameRegex)
	}
	return nil
}
