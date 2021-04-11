package services

import (
	"os"
	// "strconv"

	"github.com/streadway/amqp"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	// "github.com/buidl-labs/filecoin-chain-indexer/model"
	// messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
)

// func failOnError(err error, msg string) {
// 	if err != nil {
// 		log.Errorf("%s: %s", msg, err)
// 	}
// }

func InsertTransformedMessages(cfg config.Config) error {
	projectRoot := os.Getenv("ROOTDIR")

	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		log.Errorw("setup indexer, connecting db", "error", err)
		return xerrors.Errorf("setup indexer, connecting db: %w", err)
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q2, err := ch.QueueDeclare(
		"B_transformedcsv", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

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
		q2.Name, // queue
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
			log.Infof("Received a message: %s", msgBody)
			if msgBody[:29] == "successfully written csv file" {
				epoch := msgBody[30:]
				// epochInt, err := strconv.ParseInt(epoch, 10, 64)
				if err != nil {
					log.Error("epochInt", err)
					return
				}
				log.Info("epoch:", epoch)
				_, err = store.PqDB.Query("copy tmp_parsed_messages from '" + projectRoot + "/s3data/csvs/parsed_messages/" + epoch + ".csv' CSV HEADER")
				if err != nil {
					log.Error("copy tmp_parsed_messages", err)
					return
				}
				_, err = store.PqDB.Query("copy tmp_transactions from '" + projectRoot + "/s3data/csvs/transactions/" + epoch + ".csv' CSV HEADER")
				if err != nil {
					log.Error("copy tmp_transactions", err)
					return
				}

				// tx, err := store.PqDB.Begin()
				// if err != nil {
				// 	log.Error("db.begin", err)
				// 	return
				// }
				// _, err = tx.Exec(`copy tmp_parsed_messages from '$1/s3data/csvs/parsed_messages/$2.csv' CSV HEADER`, projectRoot, epochInt)
				// if err != nil {
				// 	log.Error("copy tmp_parsed_messages", err)
				// 	return
				// }
				// _, err = tx.Exec(`copy tmp_transactions from '$1/s3data/csvs/transactions/$2.csv' CSV HEADER`, projectRoot, epochInt)
				// if err != nil {
				// 	log.Error("copy tmp_transactions", err)
				// 	return
				// }
				// err = tx.Commit()
				// if err != nil {
				// 	log.Error("tx.commit", err)
				// 	return
				// }
				_, err = store.PqDB.Query(`INSERT INTO parsed_messages SELECT * FROM `+
					`tmp_parsed_messages WHERE height = $1 `+
					`ON CONFLICT DO NOTHING`, string(epoch))
				if err != nil {
					log.Error("insert parsed_messages", err)
					return
				}
				_, err = store.PqDB.Query(`INSERT INTO transactions SELECT * FROM `+
					`tmp_transactions WHERE height = $1 `+
					`ON CONFLICT DO NOTHING`, string(epoch))
				if err != nil {
					log.Error("insert transactions", err)
					return
				}

				// Delete inserted rows from tmp tables

				_, err = store.PqDB.Query(`DELETE FROM tmp_parsed_messages `+
					`WHERE height = $1`, string(epoch))
				if err != nil {
					log.Error("delete tmp_parsed_messages", err)
					return
				}
				_, err = store.PqDB.Query(`DELETE FROM tmp_transactions `+
					`WHERE height = $1`, string(epoch))
				if err != nil {
					log.Error("delete tmp_transactions", err)
					return
				}

				body := "successfully inserted to db " + epoch
				log.Debug(body)
				err = ch.Publish(
					"",      // exchange
					q3.Name, // routing key
					false,   // mandatory
					false,   // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(body),
					})
				failOnError(err, "Failed to publish a message")
			}
		}
	}()
	log.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	// txn:=messagemodel.Transaction{}
	// store.DB.Model()
	return nil
}
