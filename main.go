package main

import (
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Message struct {
	Payload struct {
		CommiterName string `json:"committer_name"`
		BuildingTime int    `json:"build_time_millis"`
		Branch       string `json:"branch"`
		Status       string `json:"status"`
	}
}

func main() {
	http.HandleFunc("/hooks/circle", handleCircleHook)

	port := ":" + os.Getenv("PORT")
	log.Printf("Listening on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}

	handleMessages()
}

func handleMessages() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		reply := "Don't text me"
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		switch update.Message.Command() {
		case "start":
			reply = "CircleCI Bot helps track your build status"
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}
}

func handleCircleHook(rw http.ResponseWriter, req *http.Request) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))

	var m Message
	err = json.NewDecoder(req.Body).Decode(&m)
	if err != nil {
		log.Println(err)
	}

	p := m.Payload
	statusIcon := "\xE2\x9C\x85"
	if p.Status == "failed" {
		statusIcon = "\xE2\x9D\x8C"
	}

	text := fmt.Sprintf("%s Build %s %s pushed commit to %s\n Build time: %d seconds", statusIcon, p.Status, p.CommiterName, p.Branch, p.BuildingTime/1000)
	chatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	bot.Send(tgbotapi.NewMessage(chatID, text))
}
