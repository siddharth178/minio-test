package main

import (
	"fmt"
	"github.com/minio/minio-go"
	"log"
)

func main() {
	fmt.Println("running minio test...")

	ssl := false

	mClient, err := minio.New("localhost:9000",
		"007EXR1VM2MJE7868J2K",
		"HGWm8h512+CWyl5iQDxXxOWfvfEzAHG9NEs907R6", ssl)

	if err != nil {
		log.Fatalln(err)
	}

	err = mClient.MakeBucket("mybucket", "us-east-1")
	if err != nil {
		exists, err := mClient.BucketExists("mybucket")
		if err == nil && exists {
			log.Println("we already have your bucket")
		} else {
			log.Fatalln(err)
		}
	}

	fmt.Println("Successfully created mybucket")

	n, err := mClient.FPutObject("mybucket", "myobject.txt", "/Users/siddharth/myprogs/go/src/minio-test/minio-api-client.go", "application/text")
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Successfully uploaded %s of size %d\n", "minio-api-client.go", n)

}
