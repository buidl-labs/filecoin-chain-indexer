package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	// "github.com/jasonlvhit/gocron"

	// baselens "github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/services"
)

func walk(cfg config.Config) {
	allEpochsTasks := []string{"messages", "blocks"}
	err := services.Walk(cfg, allEpochsTasks, 0) // taskType=0
	if err != nil {
		fmt.Println("services.walk err", err)
		log.Fatal(err)
	}
	currentEpochTasks := []string{"miners", "markets"}
	err = services.Walk(cfg, currentEpochTasks, 1) // taskType=1
	if err != nil {
		fmt.Println("services.walk err", err)
		log.Fatal(err)
	}

	fmt.Println("\n\n\nDONE ONE ROUND\n\n\n")
}

func main() {
	cfg := config.Config{
		DBConnStr:       "postgres://rajdeep@localhost/filecoinminermarketplace?sslmode=disable",
		FullNodeAPIInfo: os.Getenv("FULLNODE_API_INFO"),
		CacheSize:       1,
	}
	fmt.Println("Hello, world")
	// lens, err := baselens.New("rpcendpoint")
	// if err != nil {
	// 	fmt.Println("lens.New err", err)
	// 	log.Fatal(err)
	// }
	// defer lens.Close()

	var command string

	flag.StringVar(&command, "cmd", "", "Command to run")
	flag.Parse()
	if command == "" {
		log.Fatal("Command is required")
	}

	switch command {
	case "migrate", "rollback":
		services.RunMigrations(cfg, command)
	case "index":
		// minuteTicker := time.NewTicker(60 * time.Minute)
		minuteTicker := time.NewTicker(20 * time.Second)
		for {
			select {
			case <-minuteTicker.C:
				walk(cfg)
			}
		}
		// walk()
		// gocron.Every(1).Minute().Do(walk)
	default:
		log.Fatal("Please use a valid command")
	}
}
