package services

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/filecoin-project/lotus/api"
	"github.com/streadway/amqp"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	marketmodel "github.com/buidl-labs/filecoin-chain-indexer/model/market"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
	}
}

func WatchMarket(cfg config.Config) {
	projectRoot := os.Getenv("ROOTDIR")

	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		log.Errorw("setup indexer, connecting db", "error", err)
		return
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"D_marketdealsjson", // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			msgBody := string(d.Body)
			if msgBody[:36] == "successfully written marketDealsFile" {
				log.Debug("written marketDeals")

				jsonFile, err := os.Open(projectRoot + "/s3data/marketDeals/" + msgBody[37:] + ".json")
				if err != nil {
					log.Error("opening marketDealsFile", err)
					return
					// return xerrors.Errorf("opening a file in marketDeals dir: %w", err)
				}
				// defer jsonFile.Close()

				byteValue, _ := ioutil.ReadAll(jsonFile)
				jsonFile.Close()

				// var cso *api.ComputeStateOutput
				var marketDeals map[string]api.MarketDeal

				json.Unmarshal(byteValue, &marketDeals)
				// msgBody[37:] + ".json"

				for idStr, deal := range marketDeals {
					dealID, err := strconv.ParseUint(idStr, 10, 64)
					if err != nil {
						return
					}
					mds := &marketmodel.MarketDealState{
						// Height:           int64(ts.Height()),
						DealID:           dealID,
						SectorStartEpoch: int64(deal.State.SectorStartEpoch),
						LastUpdateEpoch:  int64(deal.State.LastUpdatedEpoch),
						SlashEpoch:       int64(deal.State.SlashEpoch),
						StateRoot:        "0", // ts.ParentState().String(),
					}
					mdp := &marketmodel.MarketDealProposal{
						// Height:               int64(ts.Height()),
						DealID:               dealID,
						StateRoot:            "0", // ts.ParentState().String(),
						PaddedPieceSize:      uint64(deal.Proposal.PieceSize),
						UnpaddedPieceSize:    uint64(deal.Proposal.PieceSize.Unpadded()),
						StartEpoch:           int64(deal.Proposal.StartEpoch),
						EndEpoch:             int64(deal.Proposal.EndEpoch),
						ClientID:             deal.Proposal.Client.String(),
						ProviderID:           deal.Proposal.Provider.String(),
						ClientCollateral:     deal.Proposal.ClientCollateral.String(),
						ProviderCollateral:   deal.Proposal.ProviderCollateral.String(),
						StoragePricePerEpoch: deal.Proposal.StoragePricePerEpoch.String(),
						PieceCID:             deal.Proposal.PieceCID.String(),
						IsVerified:           deal.Proposal.VerifiedDeal,
						Label:                deal.Proposal.Label,
					}

					r, err := store.DB.Model(mdp).
						OnConflict("(deal_id) DO UPDATE").
						Insert()
					if err != nil {
						log.Errorw("inserting marketDealProposal", "error", err)
					} else {
						log.Debug("inserted marketDealProposal", r)
					}
					r, err = store.DB.Model(mds).
						OnConflict("(deal_id) DO UPDATE").
						Insert()
					if err != nil {
						log.Errorw("inserting marketDealState", "error", err)
					} else {
						log.Debug("inserted marketDealState", r)
					}
				}
			}
		}
	}()
	log.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func WatchEvents(cfg config.Config) {
	// projectRoot := os.Getenv("ROOTDIR")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"A_rawcsojson", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	q2, err := ch.QueueDeclare(
		"B_transformedcsv", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			msgBody := string(d.Body)
			if msgBody[:28] == "successfully written raw cso" {
				log.Debug("written raw cso")

				/*
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
					log.Debug("fname:", projectRoot+"/s3data/cso/"+msgBody[29:]+".json")
					file, err := os.Open(projectRoot + "/s3data/cso/" + msgBody[29:] + ".json")
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
						Key:                  aws.String(msgBody[29:] + ".json"),
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
				*/

				/*** s3 upload done, start transform ***/
				err = Transform(cfg, msgBody[29:]+".json")
				if err != nil {
					log.Errorw("Transform step",
						"error", err,
						"height", msgBody[29:])
					return
				}

				// body := "failed fetching raw cso " + pts.Height().String()
				body := "successfully written csv file " + msgBody[29:]
				log.Debug(body)
				err = ch.Publish(
					"",      // exchange
					q2.Name, // routing key
					false,   // mandatory
					false,   // immediate
					amqp.Publishing{
						DeliveryMode: amqp.Persistent,
						ContentType:  "text/plain",
						Body:         []byte(body),
					})
				failOnError(err, "Failed to publish a message")

			}
			log.Infof("Received a message: %s", d.Body)
		}
	}()

	log.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	/*
		fsnotify

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

	*/
}
