package settings

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
)

const (
	ec2AccessKeyLength     = 20
	ec2SecretKeyLength     = 40
	minEc2BucketNameLength = 3
	ec2BucketNameRegex     = "^[-_a-zA-Z0-9\\.]+$"
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyEc2AccessKey(accessKey string) error {
	if len(accessKey) != ec2AccessKeyLength {
		return fmt.Errorf("Access key must be %d characters long", ec2AccessKeyLength)
	}
	return nil
}

func (verifier *Verifier) verifyEc2SecretKey(secretKey string) error {
	if len(secretKey) != ec2SecretKeyLength {
		return fmt.Errorf("Secret key must be %d characters long", ec2SecretKeyLength)
	}
	return nil
}

func (verifier *Verifier) verifyEc2BucketName(bucketName string) error {
	if len(bucketName) < minEc2BucketNameLength {
		return fmt.Errorf("Bucket name must be %d characters long", minEc2BucketNameLength)
	} else if ok, err := regexp.MatchString(ec2BucketNameRegex, bucketName); !ok || err != nil {
		return errors.New("Bucket name must match regex: " + ec2BucketNameRegex)
	}
	return nil
}
