package config

import "os"

type bucketConfig struct {
	ServiceAccountPath string
	BucketName         string
}

func (c *Config) newBucketConfig() {
	b := bucketConfig{
		ServiceAccountPath: os.Getenv("FIREBASE_BUCKET_SERVICE_ACCOUNT_PATH"),
		BucketName:         os.Getenv("FIREBASE_BUCKET_NAME"),
	}

	c.Bucket = b
}
