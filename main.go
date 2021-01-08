package main

import (
	"fmt"
	"log"

	// baselens "github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/services"
)

func main() {
	fmt.Println("Hello, world")
	// lens, err := baselens.New("rpcendpoint")
	// if err != nil {
	// 	fmt.Println("lens.New err", err)
	// 	log.Fatal(err)
	// }
	// defer lens.Close()

	err := services.Walk()
	if err != nil {
		fmt.Println("services.walk err", err)
		log.Fatal(err)
	}
}
