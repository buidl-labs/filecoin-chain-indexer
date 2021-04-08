package services

import (
	// "context"
	// "time"
	"bytes"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fsnotify/fsnotify"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
)

func WatchEvents(cfg config.Config) {
	projectRoot := os.Getenv("ROOTDIR")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Debug("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Debug("modified file:", event.Name)

					// s3 upload
					AccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
					SecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
					//  MyRegion := os.Getenv("AWS_REGION")

					sess := session.New(&aws.Config{
						Region: aws.String("us-west-2"),
						Credentials: credentials.NewStaticCredentials(
							AccessKeyID,
							SecretAccessKey,
							"", // a token will be created when the session it's used.
						)})
					// svc := s3.New(session.New(), &aws.Config{
					// 	Region: aws.String("us-west-2"),
					// 	Credentials: credentials.NewStaticCredentials(
					// 		AccessKeyID,
					// 		SecretAccessKey,
					// 		"", // a token will be created when the session it's used.
					// 	),
					// })

					file, err := os.Open(event.Name)
					if err != nil {
						log.Error("fopen", err)
						return
					}
					defer file.Close()
					// Get file size and read the file content into a buffer
					fileInfo, _ := file.Stat()
					var size int64 = fileInfo.Size()
					buffer := make([]byte, size)
					file.Read(buffer)

					_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
						Bucket:               aws.String("fmm-cso"),
						Key:                  aws.String(strings.Split(event.Name, "/")[len(strings.Split(event.Name, "/"))-1]),
						ACL:                  aws.String("private"),
						Body:                 bytes.NewReader(buffer),
						ContentLength:        aws.Int64(size),
						ContentType:          aws.String(http.DetectContentType(buffer)),
						ContentDisposition:   aws.String("attachment"),
						ServerSideEncryption: aws.String("AES256"),
					})
					if err != nil {
						log.Error("uploading to s3: ", err)
						return
					}
					log.Debug("uploaded to s3")
					Transform(cfg, strings.Split(event.Name, "/")[len(strings.Split(event.Name, "/"))-1])
					// ctx := context.Background()
					// timeout, _ := time.ParseDuration("3s")
					// var cancelFn func()
					// if timeout > 0 {
					// 	ctx, cancelFn = context.WithTimeout(ctx, timeout)
					// }
					// if cancelFn != nil {
					// 	defer cancelFn()
					// }
					// _, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
					// 	Bucket: aws.String("fmm-cso"),
					// 	Key:    aws.String(key),
					// 	Body:   os.Stdin,
					// })
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("error:", err)
			}
		}
	}()

	err = watcher.Add(projectRoot + "/s3data/cso")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
