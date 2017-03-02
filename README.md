# minio-test
Simple client that uses aws s3 apis to upload all the files from given directory to minio server running inside a container.  

Upload happens in _batches_. Batches control how many maximum concurrent uploads should be running at any given time. You can select batchsize by providing a value on command-line of client. This helps in *not* consuming
all the resources available and stressing out minio server.

## How to use it
### Start minio container as mentioned in minio doc
```
SMPro:~ siddharth$ docker run -p 9000:9000 minio/minio server /export
Created minio configuration file successfully at /root/.minio

Endpoint:  http://172.17.0.2:9000  http://127.0.0.1:9000
AccessKey: T11HKWWI3Z7ZUUOU0E8T
SecretKey: +nzKnvJtXyVgQEbOrTWmti2Q62F7dsXVd0zZMSZY
Region:    us-east-1
SQS ARNs:  <none>

Browser Access:
   http://172.17.0.2:9000  http://127.0.0.1:9000

Command-line Access: https://docs.minio.io/docs/minio-client-quickstart-guide
   $ mc config host add myminio http://172.17.0.2:9000 T11HKWWI3Z7ZUUOU0E8T +nzKnvJtXyVgQEbOrTWmti2Q62F7dsXVd0zZMSZY

Object API (Amazon S3 compatible):
   Go:         https://docs.minio.io/docs/golang-client-quickstart-guide
   Java:       https://docs.minio.io/docs/java-client-quickstart-guide
   Python:     https://docs.minio.io/docs/python-client-quickstart-guide
   JavaScript: https://docs.minio.io/docs/javascript-client-quickstart-guide

Drive Capacity: 57 GiB Free, 63 GiB Total
```
### Run aws-minio client
Use Access key and Secret key logged by minio container while running the aws-minio client. Minio container generates a new key pair everytime it is started. So do check if thats matching with your current container's values.
```
SMPro:aws-minio siddharth$ $GOPATH/bin/aws-minio -akey T11HKWWI3Z7ZUUOU0E8T -skey +nzKnvJtXyVgQEbOrTWmti2Q62F7dsXVd0zZMSZY -sourceDir . -batchSize 3
```
#### This is the log of upload activity.
```
2017/03/02 12:39:45 Uploading files from following dir: .
2017/03/02 12:39:45 AWS session created using given config
2017/03/02 12:39:45 Given bucket already exists and owned by you, continuing with it.
2017/03/02 12:39:45 Processing dir: .
2017/03/02 12:39:45 Upload worker started. Waiting for files to process...
2017/03/02 12:39:45 Sending file to worker: /Users/siddharth/myprogs/go/src/minio-test/aws-minio/aws-api-client-p.go
2017/03/02 12:39:45 Wait for all the files to be uploaded
2017/03/02 12:39:45 Processing file: /Users/siddharth/myprogs/go/src/minio-test/aws-minio/aws-api-client-p.go
2017/03/02 12:39:45 Successfully uploaded object to: mybucket  with key: /Users/siddharth/myprogs/go/src/minio-test/aws-minio/aws-api-client-p.go  and loc: http://localhost:9000/mybucket/Users/siddharth/myprogs/go/src/minio-test/aws-minio/aws-api-client-p.go
2017/03/02 12:39:45 Waited 24.387008ms for goroutines to get over
2017/03/02 12:39:45 Upload complete. Files processed: 1
2017/03/02 12:39:45 Total time required to upload all the files: 24.477791ms
```
