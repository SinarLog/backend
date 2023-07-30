package bucket

import (
	"context"
	"log"
	"sync"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var (
	once                    sync.Once
	bucketSingletonInstance *Bucket
)

var (
	_defaultAvatarPath          = "avatar"
	_defaultLeaveAttachmentPath = "leave"
	_defaultPublicLinkTemplate  = "https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&"
)

type Bucket struct {
	Handler             *storage.BucketHandle
	AvatarPath          string
	LeaveAttachmentPath string
	PublicLinkTemplate  string
}

func GetFirebaseBucket(appContext context.Context, bucketName, pathToServiceAccount string) *Bucket {
	if bucketSingletonInstance == nil {
		once.Do(func() {
			bucketSingletonInstance = &Bucket{
				AvatarPath:          _defaultAvatarPath,
				LeaveAttachmentPath: _defaultLeaveAttachmentPath,
				PublicLinkTemplate:  _defaultPublicLinkTemplate,
			}
			config := &firebase.Config{
				StorageBucket: bucketName,
			}

			opt := option.WithCredentialsFile(pathToServiceAccount)
			app, err := firebase.NewApp(appContext, config, opt)
			if err != nil {
				log.Fatalf("unable to create a new firebase app: %s", err.Error())
			}

			client, err := app.Storage(appContext)
			if err != nil {
				log.Fatalf("unable to connect firebase storage: %s", err.Error())
			}

			bucket, err := client.DefaultBucket()
			if err != nil {
				log.Fatalf("unable to connect to firebase bucket: %s", err.Error())
			}

			bucketSingletonInstance.Handler = bucket
		})
	}

	return bucketSingletonInstance
}
