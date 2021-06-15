package main

import (
	"flag"
	"fmt"
	"io"
	"log"
)

var (
	serverURL    = flag.String("serveraddr", "http://127.0.0.1:8080", "server url")
	listBotsURL  string
	createBotURL string
	runBotURL    string
	stopBotURL   string
	deleteBotURL string
)

type State int

const (
	ExecCommand State = iota
	ReadParams
)

func main() {
	flag.Parse()

	listBotsURL = fmt.Sprintf("%s/list", *serverURL)
	createBotURL = fmt.Sprintf("%s/create", *serverURL)
	runBotURL = fmt.Sprintf("%s/run", *serverURL)
	stopBotURL = fmt.Sprintf("%s/stop", *serverURL)
	deleteBotURL = fmt.Sprintf("%s/delete", *serverURL)

	currentBot := "unknown"
	//	printHelp()
	var line string
	fmt.Println("-------------------------------------------------")
	fmt.Println("current bot:", currentBot)
	fmt.Println("-------------------------------------------------")
	for {
		fmt.Printf(">>> ")
		_, err := fmt.Scanf("%s\n", &line)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("read failed", err)
		}

		switch line {
		case "create", "c":
			err = createBot()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println(err)
			}
		case "list", "l":
			getAllBots()
		case "chbot":
			fmt.Printf("select bot: ")
			_, err := fmt.Scanf("%s\n", &currentBot)
			if err == io.EOF {
				return
			}
		case "run":
			runBot(currentBot)
		case "stop":
			stopBot(currentBot)
		case "delete":
			deleteBot(currentBot)
		}

	}
}
