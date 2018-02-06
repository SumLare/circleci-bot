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

var bot, _ = tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))

func main() {
	http.HandleFunc("/hooks/circle", payloadHandler)

	port := ":" + os.Getenv("PORT")
	log.Printf("Listening on %s...\n", port)
	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Println(err)
		}
	}()

	handleMessages()
}

func handleMessages() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Println(err)
	}
	defer c.Close()

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		reply := "Don't text me"
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		chatID := update.Message.Chat.ID
		key, _ := generateKey()
		switch update.Message.Command() {
		case "start":
			reply = "CircleCI Bot helps track your build status"
		case "add":
			c.Do("APPEND", key, chatID)
			reply = "Your circleci key is " + key
		}

		msg := tgbotapi.NewMessage(chatID, reply)
		bot.Send(msg)
	}
}

func payloadHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var m Message
	err := json.NewDecoder(req.Body).Decode(&m)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Println(err)
	}
	defer c.Close()

	value, _ := c.Do("GET", req.URL.Query()["circleci_key"][0])
	chatID, _ := strconv.ParseInt(string(value.([]uint8)), 10, 64)
	sendMessage(m, chatID)
}

func sendMessage(m Message, chatID int64) {
	p := m.Payload
	statusIcon := "\xE2\x9C\x85"
	if p.Status == "failed" {
		statusIcon = "\xE2\x9D\x8C"
	}

	text := fmt.Sprintf("%s in build #%d of %s (%s) \n- %s: %s (%s) \nBuild time: %d seconds",
		statusIcon, p.BuildNumber, p.VCSUrl[8:], p.Branch, p.CommiterName,
		p.Commit, p.VCSRevision[:7], p.BuildingTime/1000)
	bot.Send(tgbotapi.NewMessage(chatID, text))
}

func generateKey() (string, error) {
	key := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, key)
	if n != len(key) || err != nil {
		return "", err
	}
	key[8] = key[8]&^0xc0 | 0x80
	key[6] = key[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", key[0:4], key[4:6], key[6:8], key[8:10], key[10:]), nil
}
