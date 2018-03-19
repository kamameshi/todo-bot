package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nlopes/slack"
)

type myEnv struct {
	BotToken string
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
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello", ev.Channel))
		}
	}

	return 0
}

func getMyEnv() myEnv {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return myEnv{
		BotToken: os.Getenv("BOT_TOKEN"),
	}
}
