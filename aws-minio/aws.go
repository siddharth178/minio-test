package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Run needed steps to create aws config, get session and create given bucket.
func doS3Setup(bucketName, accessKey, secretKey, awsServerAddr, region string, useSsl bool) (*session.Session, error) {
	bucket := aws.String(bucketName)

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey,
			secretKey,
			""),
		Endpoint:         aws.String(awsServerAddr),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(!useSsl),
		S3ForcePathStyle: aws.Bool(true),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	log.Println("AWS session created using given config")

	// create s3 client using session
	s3Client := s3.New(newSession)

	// create a new bucket using create bucket call
	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{Bucket: bucket})
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
		}
	}
	return newSession, nil
}
