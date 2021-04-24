package services

import (
	"bytes"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/streadway/amqp"
)

func UploadS3() {
	projectRoot := os.Getenv("ROOTDIR")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q3, err := ch.QueueDeclare(
		"C_insertedtodb", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q3.Name, // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			msgBody := string(d.Body)

			// s3 upload
			AccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
			SecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
			//  MyRegion := os.Getenv("AWS_REGION")

			sess, err := session.NewSession(&aws.Config{
				Region: aws.String("us-west-2"),
				Credentials: credentials.NewStaticCredentials(
					AccessKeyID,
					SecretAccessKey,
					"", // a token will be created when the session it's used.
				)})
			if err != nil {
				log.Error("newsession", err)
				return
			}
			log.Debug("msgbody:" + msgBody)
			log.Debug("fname:", projectRoot+"/s3data/cso/"+msgBody[28:]+".json")
			file, err := os.Open(projectRoot + "/s3data/cso/" + msgBody[28:] + ".json")
			if err != nil {
				log.Error("fopen", err)
				return
			}
			// defer file.Close()
			// Get file size and read the file content into a buffer
			fileInfo, _ := file.Stat()
			var size int64 = fileInfo.Size()
			buffer := make([]byte, size)
			file.Read(buffer)

			_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
				Bucket:               aws.String("fmm-cso"),
				Key:                  aws.String(msgBody[28:] + ".json"),
				ACL:                  aws.String("private"),
				Body:                 bytes.NewReader(buffer),
				ContentLength:        aws.Int64(size),
				ContentType:          aws.String(http.DetectContentType(buffer)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
			})
			if err != nil {
				log.Error("uploading to s3: ", err)
				file.Close()
				return
			}
			log.Debug("uploaded to s3")
			file.Close()
			e := os.Remove(projectRoot + "/s3data/cso/" + msgBody[28:] + ".json")
			if e != nil {
				log.Error("deleting csojson: ", err)
				return
			}

			txnsfile1, err := os.Open(projectRoot + "/s3data/csvs/transactions/" + msgBody[28:] + ".csv")
			if err != nil {
				log.Error("txns1 fopen", err)
				return
			}
			// defer txnsfile1.Close()
			// Get file size and read the file content into a buffer
			txnsfileInfo, _ := txnsfile1.Stat()
			var txnssize int64 = txnsfileInfo.Size()
			txnsbuffer := make([]byte, txnssize)
			txnsfile1.Read(txnsbuffer)

			_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
				Bucket:               aws.String("fmm-csv-transactions"),
				Key:                  aws.String(msgBody[28:] + ".csv"),
				ACL:                  aws.String("private"),
				Body:                 bytes.NewReader(txnsbuffer),
				ContentLength:        aws.Int64(txnssize),
				ContentType:          aws.String(http.DetectContentType(txnsbuffer)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
			})
			if err != nil {
				log.Error("txns1 putting to s3: ", err)
				txnsfile1.Close()
				return
			}
			log.Debug("txns1 uploaded to s3")
			txnsfile1.Close()

			e = os.Remove(projectRoot + "/s3data/csvs/transactions/" + msgBody[28:] + ".csv")
			if e != nil {
				log.Error("deleting txns1: ", err)
				return
			}

			pmsfile, err := os.Open(projectRoot + "/s3data/csvs/parsed_messages/" + msgBody[28:] + ".csv")
			if err != nil {
				log.Error("pms fopen", err)
				return
			}
			// defer pmsfile.Close()
			// Get file size and read the file content into a buffer
			pmsfileInfo, _ := pmsfile.Stat()
			var pmssize int64 = pmsfileInfo.Size()
			pmsbuffer := make([]byte, pmssize)
			pmsfile.Read(pmsbuffer)

			_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
				Bucket:               aws.String("fmm-csv-parsed-messages"),
				Key:                  aws.String(msgBody[28:] + ".csv"),
				ACL:                  aws.String("private"),
				Body:                 bytes.NewReader(pmsbuffer),
				ContentLength:        aws.Int64(pmssize),
				ContentType:          aws.String(http.DetectContentType(pmsbuffer)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
			})
			if err != nil {
				log.Error("pms putting to s3: ", err)
				pmsfile.Close()
				return
			}
			log.Debug("pms uploaded to s3")
			pmsfile.Close()
			e = os.Remove(projectRoot + "/s3data/csvs/parsed_messages/" + msgBody[28:] + ".csv")
			if e != nil {
				log.Error("deleting pms: ", err)
				return
			}
		}
	}()

	log.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
