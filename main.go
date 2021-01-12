package main

import (
	"fmt"
	"log"
	"time"

	// "github.com/jasonlvhit/gocron"

	// baselens "github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/services"
)

func walk() {
	allEpochsTasks := []string{"messages"}
	err := services.Walk(allEpochsTasks, 0) // taskType=0
	if err != nil {
		fmt.Println("services.walk err", err)
		log.Fatal(err)
	}
	currentEpochTasks := []string{"miners", "markets"}
	err = services.Walk(currentEpochTasks, 1) // taskType=1
	if err != nil {
		fmt.Println("services.walk err", err)
		log.Fatal(err)
	}

	fmt.Println("\n\n\nDONE ONE ROUND\n\n\n")
}

func main() {
	fmt.Println("Hello, world")
	// lens, err := baselens.New("rpcendpoint")
	// if err != nil {
	// 	fmt.Println("lens.New err", err)
	// 	log.Fatal(err)
	// }
	// defer lens.Close()

	// minuteTicker := time.NewTicker(60 * time.Minute)
	minuteTicker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-minuteTicker.C:
			walk()
		}
	}
	// walk()
	// gocron.Every(1).Minute().Do(walk)
}
