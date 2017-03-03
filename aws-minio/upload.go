package main

import (
	"log"
	"os"
	"sync"
)

type BatchState struct {
	sync.Mutex
	CurrCount int
	Max       int
}

func (bs *BatchState) CanUpload() bool {
	bs.Lock()
	can := bs.CurrCount < bs.Max
	bs.Unlock()

	return can
}

func (bs *BatchState) Update(count int) {
	bs.Lock()
	bs.CurrCount += count
	log.Println(bs.CurrCount)
	bs.Unlock()

}

// Upload uses fileName as Key of the object
func upload(fileName string, s3LikeStore S3LikeStore) {
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
	log.Println("Uploaded:", fileName)
}

func uploadWorker(fileNameChan <-chan string, wg *sync.WaitGroup, bs *BatchState, s3LikeStore S3LikeStore) {
	log.Println("Upload worker started. Waiting for files to process...")
	for {
		// wait for canupload to allow us
		for {
			if bs.CanUpload() {
				break
			}
		}

		fileName := <-fileNameChan

		log.Println("Processing:", fileName)
		wg.Add(1)
		bs.Update(1)
		go func() {
			upload(fileName, s3LikeStore)
			bs.Update(-1)

			wg.Done()
		}()
	}
}
