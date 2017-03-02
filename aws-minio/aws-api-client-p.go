package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func upload(fileName string, newSession *session.Session, bucket, key *string, doneChan chan int) {
	defer func() {
		doneChan <- 1
	}()

	// open the file and if its not diretory, continue with upload
	f, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	fStat, err := f.Stat()
	if err != nil {
		log.Println(err)
		return
	}
	// process the file only if its directory
	if fStat.IsDir() {
		log.Println("Skipping directory:", fileName)
		return
	}

	// store data to s3
	uploader := s3manager.NewUploader(newSession)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   f,
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		log.Println("Failed to upload data into s3 bucket:", err)
		return
	}

	log.Println("Successfully uploaded object to:", *bucket, " with key:", *key, " and loc:", result.Location)
}

func uploadWorker(fileNameChan <-chan string, wg *sync.WaitGroup, newSession *session.Session, bucket *string) {
	log.Println("Upload worker started. Waiting for files to process...")
	for {
		select {
		case fileName := <-fileNameChan:
			key := aws.String(fileName)
			log.Println("Processing file:", fileName)

			doneChan := make(chan int)
			go upload(fileName, newSession, bucket, key, doneChan)
			<-doneChan

			wg.Done()
		}
	}
}

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

// Recursively iterate through all the files in given directory
func processDirP(dirName string, fileNameChan chan string, wg *sync.WaitGroup) (fileCount int) {
	log.Println("Processing dir:", dirName)
	d, err := os.Open(dirName)
	if err != nil {
		log.Println("error in dir open:", err)
		return
	}
	defer d.Close()

	fNames, err := d.Readdirnames(0)
	if err != nil {
		log.Println("error in readdirnames:", err)
		return
	}
	for _, name := range fNames {
		fName := filepath.Join(dirName, name)
		fileCount += processFileP(fName, fileNameChan, wg)
	}
	return
}

// If current object is file, send it to uploadWorker through file name chan, for directory, just go inside that
// directory and look for more files.
func processFileP(fName string, fileNameChan chan string, wg *sync.WaitGroup) (fileCount int) {
	f, err := os.Open(fName)
	if err != nil {
		log.Println("error in open:", err)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Println("error in stat:", err)
		return
	} else {
		if stat.IsDir() {
			fileCount += processDirP(fName, fileNameChan, wg)
		} else {
			// send absolute path on channel so that s3 doesn't complain about
			// files with keys containing . or .. or something like that
			absPath, err := filepath.Abs(fName)
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Sending file to worker:", absPath)
				fileNameChan <- absPath
				wg.Add(1)
				fileCount++
			}
		}
	}
	return
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
