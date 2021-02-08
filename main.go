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
	allEpochsTasks := []string{"messages", "blocks"}
	err := services.Walk(cfg, allEpochsTasks, 0) // taskType=0
	if err != nil {
		log.Println("services.walk: allEpochsTasks", err)
	}
	currentEpochTasks := []string{"miners", "markets"}
	err = services.Walk(cfg, currentEpochTasks, 1) // taskType=1
	if err != nil {
		log.Println("services.walk: currentEpochTasks", err)
	}

	log.Info("\n\n\nDONE ONE ROUND\n\n\n")
}

func main() {
	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()
	from, _ := getenvInt("FROM")
	to, _ := getenvInt("TO")

	cfg := config.Config{
		DBConnStr:       os.Getenv("DB"),
		FullNodeAPIInfo: os.Getenv("FULLNODE_API_INFO"),
		CacheSize:       1, // TODO: Not using chain cache ATM
		From:            int64(from),
		To:              int64(to),
	}
	log.Info("Starting filecoin-chain-indexer")

	var command string

	flag.StringVar(&command, "cmd", "", "Command to run")
	flag.Parse()
	if command == "" {
		log.Fatal("Command is required")
	}

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
