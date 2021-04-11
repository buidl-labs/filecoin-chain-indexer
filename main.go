package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/rs/cors"

	"github.com/buidl-labs/filecoin-chain-indexer/chain"
	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/services"
	"github.com/buidl-labs/filecoin-chain-indexer/util"
)

var log = logging.Logger("main")

func main() {
	if err := logging.SetLogLevel("*", "debug"); err != nil {
		log.Fatalf("set log level: %w", err)
	}
	if err := logging.SetLogLevel("rpc", "error"); err != nil {
		log.Fatalf("set rpc log level: %w", err)
	}
	// if err := logging.SetLogLevel("bufbs", "error"); err != nil {
	// 	log.Fatalf("set bufbs log level: %w", err)
	// }

	go func() {
		log.Info(http.ListenAndServe(":6060", nil))
	}()

	// log.Infow("infoout of order tipsets",
	// 	"height", 3,
	// )
	// log.Info("doin nothin")
	// log.Infof("ksdjd %s", "skj")
	// log.Debugw("debugout of order tipsets", "height", 1)
	// log.Errorw("errorout of order tipsets", "height", 2)

	go func() {
		handler := cors.Default().Handler(http.FileServer(http.Dir(os.Getenv("SERVEDIR"))))
		http.Handle("/", handler)
		http.ListenAndServe(":80", handler)
	}()

	var command string
	var sfrom int
	var sto int

	fromenv, _ := util.GetenvInt("FROM")
	toenv, _ := util.GetenvInt("TO")
	flag.StringVar(&command, "cmd", "", "Command to run")
	flag.IntVar(&sfrom, "from", fromenv, "from height")
	flag.IntVar(&sto, "to", toenv, "to height")
	flag.Parse()

	if command == "" {
		log.Fatal("Please use a valid command")
	}

	epoch, _ := util.GetenvInt("EPOCH")
	forever, _ := util.GetenvInt("FOREVER")

	if sfrom == -1 {
		log.Fatal("Please specify a valid FROM epoch")
	}
	if sto == -1 && forever != 1 {
		log.Fatal("Please specify a valid TO epoch")
	}

	cfg := config.Config{
		DBConnStr:       os.Getenv("DB"),
		FullNodeAPIInfo: os.Getenv("FULLNODE_API_INFO"),
		CacheSize:       1, // TODO: Not using chain cache ATM
		From:            int64(sfrom),
		To:              int64(sto),
		Miner:           os.Getenv("MINERID"),
		Epoch:           int64(epoch),
		IndexForever:    forever,
	}
	log.Info("Starting filecoin-chain-indexer")

	switch command {
	case "migrate", "rollback":
		err := services.RunMigrations(cfg, command)
		if err != nil {
			log.Fatal("Running migrations", err)
		}
	case chain.MinerTxnsTask:
		minertxns := []string{chain.MinerTxnsTask}
		start := time.Now()
		log.Debugw("MinerTxnsTask", "start", start)
		services.Walk(cfg, minertxns, chain.EPOCHS_RANGE, chain.MinerTxnsTask)
		end := time.Now()
		log.Debugw("MinerTxnsTask", "end", end)
		diff := end.Sub(start)
		log.Debugw("MinerTxnsTask", "total time", diff)
	case chain.ActorCodesTask:
		actorcodes := []string{chain.ActorCodesTask}
		for {
			services.Walk(cfg, actorcodes, chain.SINGLE_EPOCH, chain.ActorCodesTask)
		}
	case chain.MinersTask:
		minr := []string{chain.MinersTask}
		for {
			services.Walk(cfg, minr, chain.SINGLE_EPOCH, chain.MinersTask)
		}
	case chain.BlocksTask:
		blk := []string{chain.BlocksTask}
		services.Walk(cfg, blk, chain.EPOCHS_RANGE, chain.BlocksTask)
		ticker := time.NewTicker(6 * time.Hour)
		for {
			select {
			case <-ticker.C:
				services.Walk(cfg, blk, chain.EPOCHS_RANGE, chain.BlocksTask)
			}
		}
	case chain.MessagesTask:
		msg := []string{chain.MessagesTask}
		services.Walk(cfg, msg, chain.EPOCHS_RANGE, chain.MessagesTask)
		ticker := time.NewTicker(6 * time.Hour)
		for {
			select {
			case <-ticker.C:
				services.Walk(cfg, msg, chain.EPOCHS_RANGE, chain.MessagesTask)
			}
		}
	case chain.MarketsTask:
		mkt := []string{chain.MarketsTask}
		services.Walk(cfg, mkt, chain.SINGLE_EPOCH, chain.MarketsTask)
		ticker := time.NewTicker(3 * time.Hour)
		for {
			select {
			case <-ticker.C:
				services.Walk(cfg, mkt, chain.SINGLE_EPOCH, chain.MarketsTask)
			}
		}
	// case "transform":
	// 	services.Transform(cfg)
	case "watchevents":
		services.WatchEvents(cfg)
	case "insertcsv":
		services.InsertTransformedMessages(cfg)
	default:
		log.Fatal("Please use a valid command")
	}
}
