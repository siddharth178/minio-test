package main

import (
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3LikeStore interface {
	Put(key string, body io.Reader) (string, error)
}

type minioS3Store struct {
	accessKey  string
	secretKey  string
	region     string
	serverAddr string

	bucketName string
	s3Config   *aws.Config
	session    *session.Session
	client     *s3.S3
	uploader   *s3manager.Uploader
}

// Create and return an instance tied to single bucket.
func NewMinioS3Store(accessKey, secretKey, region, serverAddr, bucketName string) (S3LikeStore, error) {
	ms := minioS3Store{accessKey: accessKey,
		secretKey:  secretKey,
		region:     region,
		serverAddr: serverAddr,
		bucketName: bucketName,
	}

	store, err := ms.initiate()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return store, nil
}

func (ms *minioS3Store) initiate() (S3LikeStore, error) {
	ms.s3Config = &aws.Config{
		Credentials: credentials.NewStaticCredentials(ms.accessKey,
			ms.secretKey,
			""),
		Endpoint:         aws.String(ms.serverAddr),
		Region:           aws.String(ms.region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}

	var err error
	ms.session, err = session.NewSession(ms.s3Config)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	log.Println("AWS session created using given config")

	// create s3 client using session
	ms.client = s3.New(ms.session)

	// create uploader
	ms.uploader = s3manager.NewUploader(ms.session)

	// create a new bucket using create bucket call
	_, err = ms.client.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(ms.bucketName)})
	if err != nil {
		// check for specific error and ignore if bucket is already exists error
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				log.Println("Given bucket already exists and owned by you, continuing with it.")
			default:
				log.Println(err)
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return ms, nil
}

func (ms *minioS3Store) Put(key string, body io.Reader) (string, error) {
	result, err := ms.uploader.Upload(&s3manager.UploadInput{
		Body:   body,
		Bucket: aws.String(ms.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Println("Failed to upload data into s3 bucket:", err)
		return "", err
	}
	return result.Location, nil
}
