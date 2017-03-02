package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

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

func isDir(fileName string) (bool, error) {
	dir, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer dir.Close()

	fInfo, err := dir.Stat()
	if err != nil {
		log.Println(err)
		return false, err
	}
	return fInfo.IsDir(), nil
}

func main() {

	// command line flags
	accessKey := flag.String("akey", "", "Access key id")
	secretKey := flag.String("skey", "", "Secret key")
	bucketName := flag.String("bucket", "mybucket", "Bucket name to be used")
	awsServerAddr := flag.String("serverAddr", "http://localhost:9000", "Aws server to connect to")
	region := flag.String("region", "us-east-1", "Aws region to use")
	useSsl := flag.Bool("ssl", false, "Use SSL while talking to aws")
	source := flag.String("sourceDir", "", "Files from this directory will get uploaded")
	batchSize := flag.Int("batchSize", 5, "Files will be uploaded to s3 in batches of this size")
	flag.Parse()

	if *accessKey == "" || *secretKey == "" || *source == "" {
		fmt.Println("Insufficient information provided")
		flag.PrintDefaults()
		return
	}

	// before making any AWS related thing, just check if given source exists
	ok, err := isDir(*source)
	if err != nil {
		log.Println(err)
		return
	}
	if !ok {
		log.Println("Given source is not a directory")
		return
	}
	log.Println("Uploading files from following dir:", *source)

	// setup s3 session and create bucket
	newSession, err := doS3Setup(*bucketName, *accessKey, *secretKey, *awsServerAddr, *region, *useSsl)
	if err != nil {
		log.Println(err)
		return
	}

	// start upload worker and make it listen on file name chan
	var wg sync.WaitGroup
	fileNameChan := make(chan string, *batchSize-1)

	// upload worker will wait on a chan for a filename
	go uploadWorker(fileNameChan, &wg, newSession, aws.String(*bucketName))

	ut1 := time.Now()
	// iterate through list of files in given directory and upload them in parallel
	// processDirP will pass in
	fileCount := processDirP(*source, fileNameChan, &wg)

	log.Println("Wait for all the files to be uploaded")
	t1 := time.Now()
	wg.Wait()
	log.Println("Waited", time.Now().Sub(t1), "for goroutines to get over")

	log.Println("Upload complete. Files processed:", fileCount)
	log.Println("Total time required to upload all the files:", time.Now().Sub(ut1))
}
