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
	BotToken   string
	BotId      string
	MongodbUrl string
}

type Todo struct {
	ID     bson.ObjectId `bson:"_id"`
	Title  string        `bson:"title"`
	Assign string        `bson:"assign"`
}

const (
	database       = "todo_schema"
	todoCollection = "todo"
)

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	env := getMyEnv()

	session, err := mgo.Dial(env.MongodbUrl)
	if err != nil {
		log.Printf("Failed open mongo: %s", err)
		return 1
	}

	defer session.Close()
	db := session.DB(database)

	api := slack.New(env.BotToken)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := handleMessageEvent(rtm, ev, env, db); err != nil {
				log.Printf("Failed to handle message: %s", err)
			}
		}
	}

	return 0
}
func addTodoList(title, assign string, db *mgo.Database) string {
	todo := &Todo{
		bson.NewObjectId(),
		title,
		assign,
	}

	if err := db.C(todoCollection).Insert(todo); err != nil {
		log.Fatalf("Failed insert: %s", err)
		return "add task failed"
	}

	return "add task completed\n" + showTodoList(db)
}

func getTitleAndAssign(text string) (string, string) {
	assign := getAssign(text)
	if len(assign) == 0 {
		// no assign
		return text, assign
	} else {
		// assigned
		m := strings.Split(text, " ")
		mExcludedAssigned := m[:len(m)-1]
		return strings.Join(mExcludedAssigned, " "), assign
	}
}

func getAssign(text string) string {
	m := strings.Split(text, " ")
	lastM := m[len(m)-1]

	if strings.HasPrefix(lastM, "<@") {
		return lastM
	} else {
		return ""
	}
}

func handleMessageEvent(rtm *slack.RTM, ev *slack.MessageEvent, env myEnv, db *mgo.Database) error {
	// response only mention
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", env.BotId)) {
		return nil
	}

	var response string
	m := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	switch m[0] {
	case "list":
		response = showTodoList(db)
	case "add":
		titleWithAssign := strings.Join(m[1:], " ")
		title, assign := getTitleAndAssign(titleWithAssign)
		response = addTodoList(title, assign, db)
	case "done", "delete":
		response = deleteTodoList(m[1], db)
	default:
		response = getHelpMessage()
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(response, ev.Channel))
	return nil
}

func showTodoList(db *mgo.Database) string {
	var todoList []Todo

	db.C(todoCollection).Find(nil).All(&todoList)
	log.Printf("data: %s", todoList)

	var result string
	for _, todo := range todoList {
		result += fmt.Sprintf("ID: %s - Task: %s - Assign: %s\n", todo.ID.Hex(), todo.Title, todo.Assign)
	}

	if len(result) == 0 {
		return "no task list"
	} else {
		return result
	}
}

func deleteTodoList(todoId string, db *mgo.Database) string {
	id := bson.ObjectIdHex(todoId)
	if err := db.C(todoCollection).RemoveId(id); err != nil {
		log.Printf("Failed delete todoList: %s", err)
		return fmt.Sprintf("Failed delete todoList: %s", err)
	} else {
		return "done|delete task completed\n" + showTodoList(db)
	}
}

func getHelpMessage() string {
	return `invalid command

@todobot list
@todobot add @mapyo buy milk
@todobot done|delete 5ab652b9c40d85a0b043aa71`
}

func getMyEnv() myEnv {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return myEnv{
		BotToken:   os.Getenv("BOT_TOKEN"),
		BotId:      os.Getenv("BOT_ID"),
		MongodbUrl: os.Getenv("MONGODB_URL"),
	}
}
