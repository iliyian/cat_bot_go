package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/buger/jsonparser"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/imroc/req"
)

func getToken() string {
	token, err := ioutil.ReadFile("token.txt")
	if err != nil {
		log.Panic("Failed to read token.txt", err)
	}
	token = bytes.Trim(token, "\"\n\r")
	return string(token)
}

func getCatKey() string {
	key, err := ioutil.ReadFile("key.txt")
	if err != nil {
		log.Print(err)
	}
	key = bytes.Trim(key, "\n\r")
	return string(key)
}

func loadJoke() [][]byte {
	joke, err := ioutil.ReadFile("jokes.txt")
	if err != nil {
		log.Panic("Failed to read jokes.")
	}
	jokes := bytes.Split(joke, []byte("\r\n\r\n"))
	log.Printf("Get %d jokes.", len(jokes))
	return jokes
}

type hit struct {
	URL     string `json:"largeImageURL"`
	ID      int    `json:"id"`
	PageURL string `json:"pageURL"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI(getToken())
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	jokes := loadJoke()
	catKey := getCatKey()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			if !update.Message.Chat.IsPrivate() {
				msg.ReplyToMessageID = update.Message.MessageID
			}
			switch update.Message.Command() {
			case "start":
				msg.ParseMode = "HTML"
				msg.Text = "这是苏联笑话大全XD!!\nmade by <a href=\"https://twitter.com/iliyian01\">@iliyian01</a>"
			case "help":
				msg.Text = "喵喵呜~~~想去苏联养猫吗喵~~~"
			case "joke":
				joke := jokes[rand.Intn(len(jokes))]
				msg.Text = string(joke)
			case "cat":
				param := req.Param{
					"key":            catKey,
					"q":              "猫",
					"lang":           "zh",
					"editors_choice": true,
					"per_page":       100,
					"pretty":         false,
					"page":           rand.Intn(3) + 1,
				}
				r, err := req.Get("https://pixabay.com/api/", param)
				if err != nil {
					log.Print(err)
				}
				q, err := r.ToBytes()
				if err != nil {
					log.Print(err)
				}
				hits, _, _, err := jsonparser.Get(q, "hits")
				if err != nil {
					log.Print(err)
				}

				var hs []hit
				err = json.Unmarshal(hits, &hs)
				if err != nil {
					log.Print(err)
				}
				idx := rand.Intn(len(hs))
				h := hs[idx]

				url, id, pageURL := h.URL, h.ID, h.PageURL
				r, err = req.Get(url)
				if err != nil {
					log.Print(err)
				}
				path := fmt.Sprintf("%d.jpg", id)
				err = r.ToFile(path)
				if err != nil {
					log.Print(err)
				}
				m := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, path)
				m.ParseMode = "HTML"
				m.Caption = fmt.Sprintf("id=<a href=\"%s\">%d</a>", pageURL, id)
				bot.Send(m)
				goto End
			default:
				msg.Text = "呀~~主人没教过我这个指令呢QAQ"
			}
			bot.Send(msg)
		End:
		}
	}
}
