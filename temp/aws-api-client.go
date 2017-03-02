package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {

	// command line flags
	accessKey := flag.String("akey", "", "access key id")
	secretKey := flag.String("skey", "", "secret key")
	bucketName := flag.String("bucket", "mybucket", "bucket name to be used")
	awsServerAddr := flag.String("serverAddr", "http://localhost:9000", "aws server to connect to")
	region := flag.String("region", "us-east-1", "aws region to use")
	useSsl := flag.Bool("ssl", false, "use SSL while talking to aws")
	flag.Parse()

	if *accessKey == "" || *secretKey == "" {
		fmt.Println("Insufficient information provided")
		flag.PrintDefaults()
		return
	}

	bucket := aws.String(*bucketName)
	key := aws.String("testobj")

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(*accessKey,
			*secretKey,
			""),
		Endpoint:         aws.String(*awsServerAddr),
		Region:           aws.String(*region),
		DisableSSL:       aws.Bool(!*useSsl),
		S3ForcePathStyle: aws.Bool(true),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Session created using given config")

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
				return
			}
		}
	}

	// store data to s3 bucket
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader("Simple object in minio server"),
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		log.Println("Failed to upload data into s3 bucket:", err)
		return
	}
	log.Println("Successfully uploaded object to:", *bucket, " with key:", *key)

	log.Println("Now downloading the same object")

	// do a get on uploaded object and save it locally.
	file, err := os.Create("testobj_local")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(newSession)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: bucket,
			Key:    key,
		})
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Downloaded file:", file.Name(), numBytes, "bytes")
}
