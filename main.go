package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nlopes/slack"
	"strings"
	"fmt"
)

type myEnv struct {
	BotToken string
	BotId string
}

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	env := getMyEnv()

	api := slack.New(env.BotToken)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := handleMessageEvent(rtm, ev, env); err != nil {
				log.Printf("Failed to handle message: %s", err)
			}
		}
	}

	return 0
}
func handleMessageEvent(rtm *slack.RTM, ev *slack.MessageEvent, env myEnv) error {
	// response only mention
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", env.BotId)) {
		return nil
	}

	rtm.SendMessage(rtm.NewOutgoingMessage("Hello", ev.Channel))
	return nil
}

func getMyEnv() myEnv {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return myEnv{
		BotToken: os.Getenv("BOT_TOKEN"),
		BotId: os.Getenv("BOT_ID"),
	}
}
