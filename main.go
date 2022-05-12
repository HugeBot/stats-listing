package main

import (
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/go-redis/redis"
)

var (
	bodyPattern string = "{server_count: @server_count@}"
)

type ShardStats struct {
	Id              int    `json:"id"`
	Status          string `json:"status"`
	GuildsCacheSize int    `json:"guildsCacheSize"`
	UsersCacheSize  int    `json:"usersCacheSize"`
	UpdatedAt       uint64 `json:"updatedAt"`
}

type StatsToPost struct {
	ServerCount int32
	ShardCount  int32
}

func GetRedisConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{})

	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	return rdb
}

func BuildBodyReader(pattern string, stats StatsToPost) io.Reader {

	pattern = strings.ReplaceAll(pattern, "@server_count@", string(stats.ServerCount))
	pattern = strings.ReplaceAll(pattern, "@shard_count@", string(stats.ShardCount))

	return strings.NewReader(pattern)
}

func GetShardStatsFromDatabase(rdb *redis.Client) []ShardStats {
	result, err := rdb.HGetAll("shard-stats").Result()
	if err != nil {
		log.Fatal(err)
	}

	stats := []ShardStats{}

	for _, value := range result {
		s := ShardStats{}
		if err := json.NewDecoder(strings.NewReader(value)).Decode(&s); err != nil {
			log.Fatal(err)
		}

		stats = append(stats, s)
	}

	return stats
}

func main() {

	rdb := GetRedisConnection()

	stats := GetShardStatsFromDatabase(rdb)

	shardCount := len(stats)
	serverCount := 0

	for _, value := range stats {
		serverCount += value.GuildsCacheSize
	}

	log.Printf("Total shards: %d, total servers: %d\n", shardCount, serverCount)

}
