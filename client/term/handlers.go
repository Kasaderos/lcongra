package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type AllBotsResponse []BotInfo

type BotInfo struct {
	ID        string  `json:"id"`
	Pair      string  `json:"pair"`
	Interval  string  `json:"interval"`
	State     string  `json:"state"`
	Cache     float64 `json:"cache"`
	Apikey    string  `json:"apikey,omitempty"`
	Apisecret string  `json:"apisecret,omitempty"`
}

func getAllBots() {
	data, err := json.Marshal(BotInfo{})
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := http.Post(listBotsURL, "", bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var bots AllBotsResponse
	err = json.Unmarshal(data, &bots)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("list:\n")
	for _, v := range bots {
		fmt.Printf("%10s %10s %5s %10s %5f\n", v.ID, v.Pair, v.Interval, v.State, v.Cache)
	}
}

func createBot() error {
	var newBot BotInfo
	var line string
	fmt.Printf("id: ")
	_, err := fmt.Scanf("%s\n", &line)
	if err != nil {
		return err
	}
	newBot.ID = line
	fmt.Printf("pair: ")
	_, err = fmt.Scanf("%s\n", &line)
	if err != nil {
		return err
	}
	newBot.Pair = line
	fmt.Printf("interval: ")
	_, err = fmt.Scanf("%s\n", &line)
	if err != nil {
		return err
	}
	newBot.Interval = line

	fmt.Printf("apikey: ")
	_, err = fmt.Scanf("%s\n", &line)
	if err != nil {
		return err
	}
	newBot.Apikey = line

	fmt.Printf("apisecret: ")
	_, err = fmt.Scanf("%s\n", &line)
	if err != nil {
		return err
	}
	newBot.Apisecret = line

	data, err := json.Marshal(newBot)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", createBotURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("create bot failed")
		return nil
	}

	return nil
}

func runBot(botID string) {
	var newBot BotInfo
	newBot.ID = botID
	data, err := json.Marshal(newBot)
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", runBotURL, bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("run bot failed")
	}
}

func stopBot(botID string) {
	var newBot BotInfo
	newBot.ID = botID
	data, err := json.Marshal(newBot)
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", stopBotURL, bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("stop bot failed")
	}
}

func deleteBot(botID string) {
	var newBot BotInfo
	newBot.ID = botID
	data, err := json.Marshal(newBot)
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", deleteBotURL, bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("delete bot failed")
	}
}
