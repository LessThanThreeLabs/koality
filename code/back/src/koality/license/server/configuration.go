package licenseserver

const upgradePathPrefix = "upgrade-v1"

type ServerConfiguration struct {
	Database DatabaseConfiguration `json:"database"`
	S3       S3Configuration       `json:"s3"`
}

type DatabaseConfiguration struct {
	Host         string `json:"host"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"databaseName"`
	SslMode      string `json:"sslMode"`
}

type S3Configuration struct {
	AwsAccessKeyId     string `json:"awsAccessKeyId"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
	BucketName         string `json:"bucketName"`
}
