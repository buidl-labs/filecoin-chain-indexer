package util

import (
	"os"
	"strconv"
)

func GetenvInt(key string) (int, error) {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return -1, err
	}
	return v, nil
}
