package miner

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/filecoin-project/go-address"
)

type ActiveMiners struct {
	ActiveMiners []string `json:"active_miners"`
}

func ActiveMinerAddresses() ([]address.Address, error) {
	// am := []string{
	// 	"f0132461", "f0395793", "f0241858",
	// 	"f0238968", "f02576", "f0224599", "f033356",
	// }
	jsonFile, err := os.Open(os.Getenv("ACTIVE_MINERS"))
	if err != nil {
		log.Errorw("loading active miners", "error", err)
	}
	defer jsonFile.Close()

	var ams ActiveMiners
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &ams)

	var ama []address.Address
	for _, addr := range ams.ActiveMiners {
		aa, _ := address.NewFromString(addr)
		ama = append(ama, aa)
	}

	// for _, a := range am {
	// 	aa, _ := address.NewFromString(a)
	// 	ama = append(ama, aa)
	// }

	return ama, nil
}
