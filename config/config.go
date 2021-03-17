package config

import (
	// goredis "github.com/go-redis/redis/v8"
	// "github.com/gomodule/redigo/redis"
	rejson "github.com/nitishm/go-rejson/v4"
)

type Config struct {
	DBConnStr       string
	FullNodeAPIInfo string
	CacheSize       int
	From            int64
	To              int64
	Miner           string
	Epoch           int64
	RH              *rejson.Handler
	IndexForever    int
}
