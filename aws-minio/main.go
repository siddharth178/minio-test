package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {

	// command line flags
	accessKey := flag.String("akey", "", "Access key id")
	secretKey := flag.String("skey", "", "Secret key")
	bucketName := flag.String("bucket", "mybucket", "Bucket name to be used")
	awsServerAddr := flag.String("serverAddr", "http://localhost:9000", "Aws server to connect to")
	region := flag.String("region", "us-east-1", "Aws region to use")
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

	// Create s3like store
	s3Like, err := NewMinioS3Store(*accessKey, *secretKey, *region, *awsServerAddr, *bucketName)
	if err != nil {
		log.Println("Failed to create s3 like store:", err)
		return
	}

	// start upload worker and make it listen on file name chan
	var wg sync.WaitGroup             // to wait for all the things to finish
	fileNameChan := make(chan string) // send file to process
	errorChan := make(chan error)

	bs := BatchState{
		CurrCount: 0,
		Max:       *batchSize,
	}

	// upload worker will wait on a chan for a filename
	go uploadWorker(fileNameChan, errorChan, &wg, &bs, s3Like)

	var errList []error
	go func() {
		for {
			err := <-errorChan
			errList = append(errList, err)
		}
	}()

	ut1 := time.Now()
	// iterate through list of files in given directory and upload them in parallel
	// processDirP will pass in
	fileCount := processDirP(*source, fileNameChan)

	log.Println("Wait for all the files to be uploaded")
	t1 := time.Now()
	wg.Wait()
	log.Println("Waited", time.Now().Sub(t1), "for goroutines to get over")

	log.Println("Upload complete. Files processed:", fileCount)
	log.Println("Total time required to upload all the files:", time.Now().Sub(ut1))

	log.Println(len(errList), "uploads failed")
	for _, e := range errList {
		log.Println("Err:", e)
	}
}
