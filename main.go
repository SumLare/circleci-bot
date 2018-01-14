package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/garyburd/redigo/redis"
	"io"
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
		Commit       string `json:"subject"`
		BuildNumber  int    `json:"build_num"`
		VCSUrl       string `json:"vcs_url"`
		VCSRevision  string `json:"vcs_revision"`
	}
}

func main() {
	http.HandleFunc("/hooks/circle", payloadHandler)

	port := ":" + os.Getenv("PORT")
	log.Printf("Listening on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Println(err)
	}

	handleMessages()
}

func handleMessages() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Println(err)
	}

	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Println(err)
	}
	defer c.Close()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		reply := "Don't text me"
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		chatID := update.Message.Chat.ID
		switch update.Message.Command() {
		case "start":
			reply = "CircleCI Bot helps track your build status"
		case "add":
			key, _ := newUUID()
			c.Do("APPEND", chatID, key)
			reply = "Your circleci key is " + key
		}

		msg := tgbotapi.NewMessage(chatID, reply)
		bot.Send(msg)
	}
}

func payloadHandler(rw http.ResponseWriter, req *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var m Message
	err := json.NewDecoder(req.Body).Decode(&m)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sendMessage(m)
}

func sendMessage(m Message) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Println(err)
	}

	p := m.Payload
	statusIcon := "\xE2\x9C\x85"
	if p.Status == "failed" {
		statusIcon = "\xE2\x9D\x8C"
	}

	text := fmt.Sprintf("%s in build #%d of %s (%s) \n- %s: %s (%s) \nBuild time: %d seconds",
		statusIcon, p.BuildNumber, p.VCSUrl[8:], p.Branch, p.CommiterName,
		p.Commit, p.VCSRevision[:7], p.BuildingTime/1000)
	chatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	bot.Send(tgbotapi.NewMessage(chatID, text))
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
