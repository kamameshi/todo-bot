package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nlopes/slack"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"strings"
)

type myEnv struct {
	BotToken string
	BotId    string
}

type Todo struct {
	ID     bson.ObjectId `bson:"_id"`
	Title  string        `bson:"title"`
	Assign string        `bson:"assign"`
}

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	session, err := mgo.Dial("mongodb://localhost/test")
	if err != nil {
		log.Printf("Failed open mongo: %s", err)
		return 1
	}
	defer session.Close()
	db := session.DB("todo_schema")

    todo := &Todo{
        bson.NewObjectId(),
        "test todo",
        "@mapyo",
    }

    db.C("todo").Insert(todo)

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

	var response string
	m := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	switch m[0] {
	case "list":
		response = "will show todo list"
	case "add":
		response = "the task will be add"
	case "done", "delete":
		response = "the task will be delete"
	default:
		response = "will show help"
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(response, ev.Channel))
	return nil
}

func getMyEnv() myEnv {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return myEnv{
		BotToken: os.Getenv("BOT_TOKEN"),
		BotId:    os.Getenv("BOT_ID"),
	}
}
