package pools

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"regexp"
)

const (
	minNameLength            = 1
	maxNameLength            = 256
	minUsernameLength        = 1
	maxUsernameLength        = 256
	minNumMaxInstances       = 1
	ec2AccessKeyLength       = 20
	ec2SecretKeyLength       = 40
	ec2BaseAmiIdLegth        = 12
	ec2SecurityGroupIdLength = 11
	ec2VpcSubnetIdLength     = 15
	ec2MinRootDriveSize      = 10
	nameRegex                = "^[-_a-zA-Z0-9 ]+$"
	ec2BaseAmiIdRegex        = "^ami-[a-f0-9]+$"
	ec2SecurityGroupIdRegex  = "^sg-[a-f0-9]+$"
	ec2VpcSubnetIdRegex      = "^subnet-[a-f0-9]+$"
)

var (
	allowedEc2InstanceTypes []string = []string{"m1.small", "m1.medium", "m1.large", "m1.xlarge",
		"m2.xlarge", "m2.2xlarge", "m2.4xlarge", "cr1.8xlarge",
		"m3.xlarge", "m3.2xlarge",
		"c1.medium", "c1.xlarge", "cc2.8xlarge",
		"c3.large", "c3.xlarge", "c3.2xlarge", "c3.4xlarge", "c3.8xlarge",
		"hi1.4xlarge", "hi1.8xlarge"}
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyName(name string) error {
	if len(name) < minNameLength {
		return fmt.Errorf("Name must be at least %d characters long", minNameLength)
	} else if len(name) > maxNameLength {
		return fmt.Errorf("Name cannot exceed %d characters long", maxNameLength)
	} else if ok, err := regexp.MatchString(nameRegex, name); !ok || err != nil {
		return errors.New("Name must match regex: " + nameRegex)
	} else if err := verifier.verifyPoolDoesNotExistWithName(name); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyUsername(username string) error {
	if len(username) < minUsernameLength {
		return fmt.Errorf("Username must be at least %d characters long", minUsernameLength)
	} else if len(username) > maxUsernameLength {
		return fmt.Errorf("Username cannot exceed %d characters long", maxUsernameLength)
	}
	return nil
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

func (verifier *Verifier) verifyEc2BaseAmiId(baseAmiId string) error {
	if len(baseAmiId) != ec2BaseAmiIdLegth {
		return fmt.Errorf("Base AMI Id must be %d characters long", ec2BaseAmiIdLegth)
	} else if ok, err := regexp.MatchString(ec2BaseAmiIdRegex, baseAmiId); !ok || err != nil {
		return errors.New("Base AMI Id must match regex: " + ec2BaseAmiIdRegex)
	}
	return nil
}

func (verifier *Verifier) verifyEc2SecurityGroupId(securityGroupId string) error {
	if len(securityGroupId) != ec2SecurityGroupIdLength {
		return fmt.Errorf("Security Group Id must be %d characters long", ec2SecurityGroupIdLength)
	} else if ok, err := regexp.MatchString(ec2SecurityGroupIdRegex, securityGroupId); !ok || err != nil {
		return errors.New("Security Group Id must match regex: " + ec2SecurityGroupIdRegex)
	}
	return nil
}

func (verifier *Verifier) verifyEc2VpcSubnetId(vpcSubnetId string) error {
	if len(vpcSubnetId) != ec2VpcSubnetIdLength {
		return fmt.Errorf("VPC Subnet Id must be %d characters long", ec2VpcSubnetIdLength)
	} else if ok, err := regexp.MatchString(ec2VpcSubnetIdRegex, vpcSubnetId); !ok || err != nil {
		return errors.New("VPC Subnet Id must match regex: " + ec2VpcSubnetIdRegex)
	}
	return nil
}

func (verifier *Verifier) verifyEc2InstanceType(instanceType string) error {
	for _, allowedEc2InstanceType := range allowedEc2InstanceTypes {
		if instanceType == allowedEc2InstanceType {
			return nil
		}
	}
	return fmt.Errorf("Instance type must be one of: %v", allowedEc2InstanceTypes)
}

func (verifier *Verifier) verifyReadyAndMaxInstances(numReadyInstances, numMaxInstances uint64) error {
	if numMaxInstances < minNumMaxInstances {
		return fmt.Errorf("Number max instances must be at least %d", minNumMaxInstances)
	} else if numReadyInstances > numMaxInstances {
		return errors.New("Number ready instances cannot exceed max instances")
	}
	return nil
}

func (verifier *Verifier) verifyEc2RootDriveSize(rootDriveSize uint64) error {
	if rootDriveSize < ec2MinRootDriveSize {
		return fmt.Errorf("Root drive size must be at least %dGB", ec2MinRootDriveSize)
	}
	return nil
}

func (verifier *Verifier) verifyPoolDoesNotExistWithName(name string) error {
	query := "SELECT id FROM ec2_pools WHERE name=$1"
	err := verifier.database.QueryRow(query, name).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		return resources.PoolAlreadyExistsError{"Pool already exists with name: " + name}
	}
	return nil
}
