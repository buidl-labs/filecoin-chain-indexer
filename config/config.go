package config

type Config struct {
	DBConnStr       string
	FullNodeAPIInfo string
	CacheSize       int
	From            int64
	To              int64
}
