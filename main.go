package main

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

type ShardStats struct {
	Id              int    `json:"id"`
	Status          string `json:"status"`
	GuildsCacheSize int    `json:"guildsCacheSize"`
	UsersCacheSize  int    `json:"usersCacheSize"`
	UpdatedAt       uint64 `json:"updatedAt"`
}

func GetRegisConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{})

	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	return rdb
}

func GetShardStatsFromDatabase(rdb *redis.Client) map[int]ShardStats {
	result, err := rdb.HGetAll("shard-stats").Result()
	if err != nil {
		log.Fatal(err)
	}

	var stats map[int]ShardStats

	for key, value := range result {
		var s ShardStats
		if err := json.NewDecoder(strings.NewReader(value)).Decode(s); err != nil {
			log.Fatal(err)
		}

		id, err := strconv.ParseInt(key, 0, 0)
		if err != nil {
			log.Fatal(err)
		}

		stats[int(id)] = s
	}

	return stats
}

func main() {

	rdb := GetRegisConnection()

	stats := GetShardStatsFromDatabase(rdb)

}
