package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mrshankly/go-twitch/twitch"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Text        string        `json:"text"`
	Attachments []*Attachment `json:"attachments"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback  string   `json:"fallback"`
	Title     string   `json:"title"`
	TitleLink string   `json:"title_link"`
	Text      string   `json:"text"`
	ImageUrl  string   `json:"image_url"`
	Fields    []*Field `json:"fields"`
	Color     string   `json:"color"`
}

type Attachments struct {
	Attachs []*Attachment `json:"attachments"`
}

// on behalf of whg i need to apologize for doing this so quick and dirty
// enjoy ^_^

// bot keeps running if errors, except if listenandserve fails

func Run(port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		text := r.PostFormValue("text")
		userId := r.PostFormValue("user_id")
		userName := r.PostFormValue("user_name")

		log.Printf("Handling request: %s from: %s (ID: %s)", text, userName, userId)

		// make sure we dont trigger from ourselves, not sure if whgbot is needed here though, seems like slack uses USLACKBOT anyway
		if userId == "USLACKBOT" || userName == "whgbot" {
			return
		}

		// need to check to see if the url we are fed is a normal url
		//  or if this has been altered to slack format
		slackRe := regexp.MustCompile(`([\S]+)\|(www\.[\S]+)$`)
		isSlack := slackRe.MatchString(text)

		// if it's the weird slack format, we need to get the proper url
		if isSlack {
			text = slackRe.FindStringSubmatch(text)[2]
		}

		// thank you based slack devs for making me do this
		text = strings.Replace(text, ">", "", -1)
		text = strings.Replace(text, "<", "", -1)

		// matches twitch.tv and justin.tv urls
		//  (although justin.tv is gone now?)
		streamRe := regexp.MustCompile(`(twitch\.tv\/([\w]+)|justin\.tv\/([\w]+))\/?$`)
		videoRe := regexp.MustCompile(`(twitch\.tv\/([\w]+)\/([a-z]\/[0-9]+)\/?|justin\.tv\/([\w]+)\/([a-z]\/[0-9]+)\/?)`)
		isStream := streamRe.MatchString(text)
		isVideo := videoRe.MatchString(text)

		client := twitch.NewClient(&http.Client{})

		if isStream {
			username := streamRe.FindStringSubmatch(text)[2]

			data, err := client.Streams.Channel(username)
			if err != nil {
				log.Println(err)
				return
			}

			if data.Stream.Id == 0 {
				log.Println("STREAM OFFLINE")
				response := &Response{Attachments: nil, Text: username + " is not streaming."}

				payload, err := json.Marshal(response)
				if err != nil {
					return
				}

				log.Printf("Sending response: %s", payload)

				w.Write(payload)
				return
			}

			viewers := &Field{
				Title: "Viewers",
				Value: strconv.Itoa(data.Stream.Viewers),
				Short: true,
			}
			// an attachment may contain several fields, but only one here
			fields := []*Field{viewers}

			attachment := &Attachment{
				Fallback:  username + " playing " + data.Stream.Game + " with " + strconv.Itoa(data.Stream.Viewers) + " viewers.",
				Title:     data.Stream.Channel.Status,
				TitleLink: data.Stream.Channel.Url,
				Text:      "Playing " + data.Stream.Game,
				ImageUrl:  data.Stream.Preview,
				Fields:    fields,
				Color:     "#6441A5",
			}

			// a response may contain several attachments
			attachs := []*Attachment{attachment}

			response := &Response{Attachments: attachs, Text: data.Stream.Channel.Name + " is streaming right now!"}

			payload, err := json.Marshal(response)
			if err != nil {
				return
			}

			log.Printf("Sending response: %s", payload)

			w.Write(payload)

		} else if isVideo {
			videoId := videoRe.FindStringSubmatch(text)[3]
			videoId = strings.Replace(videoId, "/", "", -1)
			data, err := client.Videos.Id(videoId)
			if err != nil {
				return
			}

			if data == nil {
				return
			}

			views := &Field{
				Title: "Views",
				Value: strconv.Itoa(data.Views),
				Short: true,
			}
			duration := time.Duration(data.Length) * time.Second
			length := &Field{
				Title: "Duration",
				Value: duration.String(),
				Short: true,
			}
			// an attachment may contain several fields, but only one here
			fields := []*Field{views, length}

			attachment := &Attachment{
				Fallback:  data.Title + " [ " + strconv.Itoa(data.Views) + " views]",
				Title:     data.Title,
				TitleLink: data.Url,
				Text:      data.Description,
				ImageUrl:  data.Preview,
				Fields:    fields,
				Color:     "#6441A5",
			}

			// a response may contain several attachments
			attachs := []*Attachment{attachment}

			response := &Response{Attachments: attachs, Text: "Video by " + data.Channel.Name}

			payload, err := json.Marshal(response)
			if err != nil {
				return
			}

			log.Printf("Sending response: %s", payload)

			w.Write(payload)
		} else {
			log.Println("IS NOTHING")
			return
		}
	})

	log.Printf("Starting http server on port %d", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	var httpPort int
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: whgbot-slack.exe -port=3000\n")
		flag.PrintDefaults()
	}
	flag.IntVar(&httpPort, "port", 27015, "The HTTP port on which to listen")

	flag.Parse()

	if httpPort == 0 {
		flag.Usage()
		os.Exit(2)
	}

	Run(httpPort)
}
