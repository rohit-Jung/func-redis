package config

var (
	Host      string = "127.0.0.1"
	Port      int    = 7379
	KeysLimit int    = 5

	AofFileName string = "aof-dump.aof"

	EvictionStrategy string  = "all-keys-random"
	EvictionRatio   float64 = 0.40
)
