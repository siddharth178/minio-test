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
	log.Println(bs.CurrCount, "\tactive")
	bs.Unlock()

}

// Upload uses fileName as Key of the object
func upload(fileName string, s3LikeStore S3LikeStore) error {
	// open the file and if its not diretory, continue with upload
	f, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	fStat, err := f.Stat()
	if err != nil {
		log.Println(err)
		return err
	}
	// process the file only if its directory
	if fStat.IsDir() {
		log.Println("Skipping directory:", fileName)
		return nil
	}

	// store data to s3
	_, err = s3LikeStore.Put(fileName, f)
	if err != nil {
		log.Println("Failed to upload data into s3 bucket:", err)
		return err
	}
	log.Println("Uploaded:", fileName)
	return nil
}

func uploadWorker(fileNameChan <-chan string, errorChan chan error, wg *sync.WaitGroup, bs *BatchState, s3LikeStore S3LikeStore) {
	log.Println("Upload worker started. Waiting for files to process...")
	for {
		// wait for canupload to allow us
		for {
			if bs.CanUpload() {
				break
			}
		}

		fileName := <-fileNameChan

		wg.Add(1)
		bs.Update(1)
		go func() {
			err := upload(fileName, s3LikeStore)
			if err != nil {
				errorChan <- err
			}
			bs.Update(-1)

			wg.Done()
		}()
	}
}
