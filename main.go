package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/services"
)

func walk(cfg config.Config) {
	// allEpochsTasks := []string{"messages", "blocks"}
	// err := services.Walk(cfg, allEpochsTasks, 0) // taskType=0
	// if err != nil {
	// 	log.Println("services.walk: allEpochsTasks", err)
	// }
	// currentEpochTasks := []string{"miners", "markets"}
	// err = services.Walk(cfg, currentEpochTasks, 1) // taskType=1
	// if err != nil {
	// 	log.Println("services.walk: currentEpochTasks", err)
	// }

	// initial := []string{"init"}
	msg := []string{"messages"}
	blk := []string{"blocks"}
	// minr := []string{"miners"}
	mkt := []string{"markets"}
	// err := services.Walk(cfg, initial, 1)
	// if err != nil {
	// 	log.Println("init task", err)
	// }
	go services.Walk(cfg, msg, 0) // taskType=0
	go services.Walk(cfg, blk, 0) // taskType=0
	// go services.Walk(cfg, minr, 1) // taskType=1
	go services.Walk(cfg, mkt, 1) // taskType=1

	log.Info("\n\n\nDONE ONE ROUND\n\n\n")
}

func main() {
	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()
	// from, _ := getenvInt("FROM")
	// to, _ := getenvInt("TO")

	// cfg := config.Config{
	// 	DBConnStr:       os.Getenv("DB"),
	// 	FullNodeAPIInfo: os.Getenv("FULLNODE_API_INFO"),
	// 	CacheSize:       1, // TODO: Not using chain cache ATM
	// 	From:            int64(from),
	// 	To:              int64(to),
	// }
	// log.Info("Starting filecoin-chain-indexer")

	var command string
	var sfrom int
	var sto int

	flag.StringVar(&command, "cmd", "", "Command to run")
	flag.IntVar(&sfrom, "from", 0, "from height")
	flag.IntVar(&sto, "to", 500000, "to height")
	flag.Parse()
	if command == "" {
		log.Fatal("Command is required")
	}
	epoch, _ := getenvInt("EPOCH")
	cfg := config.Config{
		DBConnStr:       os.Getenv("DB"),
		FullNodeAPIInfo: os.Getenv("FULLNODE_API_INFO"),
		CacheSize:       1, // TODO: Not using chain cache ATM
		From:            int64(sfrom),
		To:              int64(sto),
		Miner:           os.Getenv("MINERID"),
		Epoch:           int64(epoch),
	}
	log.Info("Starting filecoin-chain-indexer")

	switch command {
	case "migrate", "rollback":
		err := services.RunMigrations(cfg, command)
		if err != nil {
			log.Fatal("Running migrations", err)
		}
	case "index":
		walk(cfg)
		// minuteTicker := time.NewTicker(60 * time.Second)
		minuteTicker := time.NewTicker(24 * time.Hour)
		dayTicker := time.NewTicker(24 * time.Hour)
		for {
			select {
			case <-minuteTicker.C:
				walk(cfg)
			case <-dayTicker.C:
				getDailyData(cfg)
			}
		}
	case "minerinit":
		minr := []string{"miners"}
		services.Walk(cfg, minr, 1)
	case "minertxns":
		minertxns := []string{"minertxns"}
		initt := time.Now()
		log.Info("initial: ", initt)
		services.Walk(cfg, minertxns, 0)
		finall := time.Now()
		log.Info("final: ", finall)
		diff := finall.Sub(initt)
		log.Info("diff: ", diff)
	case "minerinfo":
		minerinfo := []string{"minerinfo"}
		services.Walk(cfg, minerinfo, 1)
	case "actorcodes":
		minerinfo := []string{"minerinfo"}
		services.Walk(cfg, minerinfo, 1)
		dayTicker := time.NewTicker(24 * time.Hour)
		for {
			select {
			case <-dayTicker.C:
				services.Walk(cfg, minerinfo, 1)
			}
		}
	default:
		log.Fatal("Please use a valid command")
	}
}

func getenvInt(key string) (int, error) {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return -1, err
	}
	return v, nil
}

func getDailyData(cfg config.Config) {

}
