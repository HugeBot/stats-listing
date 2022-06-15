package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/go-redis/redis"
	"gopkg.in/yaml.v2"
)

var (
	config Config

	filePath string

	ErrNoConfig = errors.New("")
)

const (
	defaultBodyPattern = "{\"server_count\": @server_count@}"
)

type Config struct {
	BotID    string       `yaml:"botId"`
	Websites []Website    `yaml:"websites"`
	Redis    RedisCondfig `yaml:"redis"`
}

type RedisCondfig struct {
	Host     string `yaml:"host"`
	Password string `yaml:"pass"`
	Port     int    `yaml:"port"`
	Db       int    `yaml:"db"`
}

type Website struct {
	Name        string `yaml:"name"`
	ApiPath     string `yaml:"apiPath"`
	Token       string `yaml:"token"`
	BodyPattern string `yaml:"bodyPattern"`
	Method      string `yaml:"method"`
}

type ShardStats struct {
	Id              int    `json:"id"`
	Status          string `json:"status"`
	GuildsCacheSize int    `json:"guildsCacheSize"`
	UsersCacheSize  int    `json:"usersCacheSize"`
	UpdatedAt       uint64 `json:"updatedAt"`
}

type Stats struct {
	ServerCount int
	ShardCount  int
}

func GetRedisConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.Db,
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	return rdb
}

func BuildBodyReader(stats Stats, website Website) io.Reader {

	var pattern string
	if len(website.BodyPattern) == 0 {
		pattern = defaultBodyPattern
	} else {
		pattern = website.BodyPattern
	}

	pattern = strings.ReplaceAll(pattern, "@server_count@", strconv.Itoa(stats.ServerCount))
	pattern = strings.ReplaceAll(pattern, "@shard_count@", strconv.Itoa(stats.ShardCount))

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

func init() {
	flag.StringVar(&filePath, "config", "./config.yaml", "Path to the config file")
	flag.Parse()

	filename, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	if len(config.BotID) == 0 {
		log.Fatal("botId must be defined con config.yaml")
	}

	if len(config.Redis.Host) == 0 {
		config.Redis.Host = "localhost"
	}

	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}

	for i, website := range config.Websites {
		if website.Name == "" {
			log.Fatal("Must define a name to website")
		}

		if website.ApiPath == "" {
			log.Fatalf("Must define a apiPath for website %s", website.Name)
		}

		if website.Token == "" {
			log.Fatalf("Must define a token for website %s", website.Name)
		}

		if website.Method == "" {
			config.Websites[i].Method = "POST"
		}

		config.Websites[i].ApiPath = strings.ReplaceAll(website.ApiPath, "@bot_id@", config.BotID)
	}
}

func PostStatsToWebsite(wg *sync.WaitGroup, stats Stats, website Website) {
	defer wg.Done()

	var req *http.Request

	if strings.Contains(website.ApiPath, "@server_count@") {
		website.ApiPath = strings.ReplaceAll(website.ApiPath, "@server_count@", fmt.Sprint(stats.ServerCount))

		r, err := http.NewRequest(strings.ToUpper(website.Method), website.ApiPath, nil)
		if err != nil {
			log.Fatal(err)
		}

		req = r
	} else {
		body := BuildBodyReader(stats, website)

		r, err := http.NewRequest(strings.ToUpper(website.Method), website.ApiPath, body)
		if err != nil {
			log.Fatal(err)
		}

		r.Header.Set("Content-Type", "application/json")

		req = r
	}

	req.Header.Add("Authorization", website.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Printf("• %s (%s): %s\n", website.Name, website.ApiPath, resp.Status)
}

func main() {

	var wg sync.WaitGroup

	log.Printf("Connecting with redis server at %s:%d\n", config.Redis.Host, config.Redis.Port)
	rdb := GetRedisConnection()

	log.Print("Retrieving HASH MAP with key shard-stats\n")
	stats := GetShardStatsFromDatabase(rdb)

	log.Print("Closing connection with the database.\n")
	rdb.Close()

	body := Stats{
		ShardCount:  len(stats),
		ServerCount: 0,
	}

	for _, value := range stats {
		body.ServerCount += value.GuildsCacheSize
	}

	log.Printf("Total shards: %d, total servers: %d\n", body.ShardCount, body.ServerCount)

	if len(config.Websites) < 1 {
		log.Fatal("No websites on config")
	}

	for _, website := range config.Websites {
		wg.Add(1)
		go PostStatsToWebsite(&wg, body, website)
	}

	wg.Wait()
}
