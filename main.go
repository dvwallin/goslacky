package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/asdine/storm"
	"github.com/rs/xid"
)

type SlackMessage struct {
	ID          string `storm:"id"`
	Token       string
	TeamID      string
	TeamDomain  string
	ChannelID   string
	ChannelName string
	Timestamp   float64
	UserID      string
	UserName    string
	Text        string
	TriggerWord string
}

func main() {

	db, err := storm.Open("goslacky.db")
	if err != nil {
		fmt.Printf("error opening database (goslacky.db): %s\n", err.Error())
		os.Exit(1)
	}

	defer db.Close()

	if len(os.Args[1:]) < 1 {
		usage()
		os.Exit(1)
	}

	if len(os.Args[1]) < 22 || len(os.Args[1]) > 26 {
		fmt.Println("the token you supplied seems to be of the wrong format")
		os.Exit(1)
	}

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		var messages []SlackMessage
		err := db.All(&messages)
		if err != nil {
			fmt.Printf("error getting all messages from database: %s\n", err.Error())
			return
		}
		t, _ := template.ParseFiles("template.html")
		t.Execute(w, messages)
	})

	http.HandleFunc("/sink", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("token") == os.Args[1] {
			guid := xid.New()
			timestamp, err := strconv.ParseFloat(r.FormValue("timestamp"), 64)
			if err != nil {
				fmt.Printf("error converting timestamp string to float64 value: %s\n", err.Error())
				return
			}
			DataObject := SlackMessage{
				ID:          guid.String(),
				Token:       r.FormValue("token"),
				TeamID:      r.FormValue("team_id"),
				TeamDomain:  r.FormValue("team_domain"),
				ChannelID:   r.FormValue("channel_id"),
				ChannelName: r.FormValue("channel_name"),
				Timestamp:   timestamp,
				UserID:      r.FormValue("user_id"),
				UserName:    r.FormValue("user_name"),
				Text:        r.FormValue("text"),
				TriggerWord: r.FormValue("trigger_word"),
			}
			saveErr := db.Save(&DataObject)
			if saveErr != nil {
				fmt.Printf("error saving DataObject to database: %s\n", err.Error())
				return
			}
		}
	})

	http.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	fmt.Println("running server on port 4444")
	http.ListenAndServe("0.0.0.0:4444", nil)
}

func usage() {
	fmt.Println("please supply the Outgoing Hook Token as an argument. ./goslacky TOKENID")
}
