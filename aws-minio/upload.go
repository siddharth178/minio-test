package main

import (
	"log"
	"os"
	"sync"
)

// Upload uses fileName as Key of the object
func upload(fileName string, s3LikeStore S3LikeStore, doneChan chan int) {
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
	_, err = s3LikeStore.Put(fileName, f)
	if err != nil {
		log.Println("Failed to upload data into s3 bucket:", err)
		return
	}

	// log.Println("Uploaded object with key:", fileName)
}

func uploadWorker(fileNameChan <-chan string, wg *sync.WaitGroup, s3LikeStore S3LikeStore, bucket string) {
	log.Println("Upload worker started. Waiting for files to process...")
	for {
		select {
		case fileName := <-fileNameChan:
			doneChan := make(chan int)
			go upload(fileName, s3LikeStore, doneChan)
			<-doneChan

			wg.Done()
		}
	}
}
