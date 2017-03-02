package main

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
