package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	config Config

	filePath   string
	shardCount int
	verbose    bool

	ErrNoConfig = errors.New("")

	Version = "0.3.3"
)

const (
	defaultBodyPattern = "{\"guild_count\": @guild_count@}"
)

type Config struct {
	BotToken string    `yaml:"botToken"`
	Websites []Website `yaml:"websites"`
}

type Website struct {
	Name        string `yaml:"name"`
	ApiPath     string `yaml:"apiPath"`
	Token       string `yaml:"token"`
	BodyPattern string `yaml:"bodyPattern"`
	Method      string `yaml:"method"`
}

type ApplicationResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	GuildCount *int   `json:"approximate_guild_count"`
}

func buildBodyReader(website Website, application ApplicationResponse) io.Reader {

	var pattern string
	if len(website.BodyPattern) == 0 {
		pattern = defaultBodyPattern
	} else {
		pattern = website.BodyPattern
	}

	pattern = strings.ReplaceAll(pattern, "@server_count@", strconv.Itoa(*application.GuildCount))
	pattern = strings.ReplaceAll(pattern, "@shard_count@", strconv.Itoa(shardCount))

	return strings.NewReader(pattern)
}

func init() {
	flag.StringVar(&filePath, "config", "./config.yaml", "Path to the config file.")
	flag.IntVar(&shardCount, "shards", 0, "The shard count.")
	flag.BoolVar(&verbose, "verbose", false, "Print Application info.")
	flag.Parse()

	filename, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
	}

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	if len(config.BotToken) == 0 {
		log.Fatal("botToken mus be defined on config.yaml")
	}

	if len(config.Websites) < 1 {
		log.Fatal("No websites on config")
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
	}
}

func postStatsToWebsite(wg *sync.WaitGroup, website Website, application ApplicationResponse) {
	defer wg.Done()

	var req *http.Request

	if strings.Contains(website.ApiPath, "@guild_count@") {
		website.ApiPath = strings.ReplaceAll(website.ApiPath, "@guild_count@", fmt.Sprint(application.GuildCount))
		website.ApiPath = strings.ReplaceAll(website.ApiPath, "@bot_id@", application.ID)

		r, err := http.NewRequest(strings.ToUpper(website.Method), website.ApiPath, nil)
		if err != nil {
			log.Fatal(err)
		}

		req = r
	} else {
		body := buildBodyReader(website, application)

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

func getApplicationInfo(botToken string) (ApplicationResponse, error) {
	var body ApplicationResponse

	req, err := http.NewRequest("GET", "https://discord.com/api/v10/applications/@me", nil)
	if err != nil {
		return body, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", botToken))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return body, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return body, err
	}

	return body, nil

}

func main() {

	var wg sync.WaitGroup

	application, err := getApplicationInfo(config.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Application info:\n")
	log.Printf("• ID: %s", application.ID)
	log.Printf("• Name: %s", application.Name)
	log.Printf("• Guild count: %d", *application.GuildCount)

	if verbose {
		return
	}

	for _, website := range config.Websites {
		wg.Add(1)
		go postStatsToWebsite(&wg, website, application)
	}

	wg.Wait()
}
