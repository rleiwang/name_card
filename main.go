package main

import (
	"flag"
	"log"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
)

var addr = flag.String("addr", ":8080", "http service address")

// NameList name list
type NameList struct {
	// Id index id
	IDX int `json:"idx"`

	UID int `json:"uid"`

	// FID family id
	FID int `json:"fid"`

	// Family name
	Family string `json:"family"`

	// ChineseName name
	ChineseName string `json:"cname"`

	// first name
	First string `json:"first"`

	// last name
	Last string `json:"last"`

	// Absent absent
	Absent bool `json:"absent"`
}

// Message message
type Message struct {
	Type string     `json:"type"`
	List []NameList `json:"list"`
}

type NameLists struct {
	lists []NameList
}

func main() {
	flag.Parse()
	b := newBroker()
	defer b.stop()
	go b.run()

	l := NameLists{lists: loadFileIfExists()}

	http.Handle("/", http.FileServer(rice.MustFindBox("build").HTTPBox()))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(b, &l, w, r)
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		handleUpload(&l, w, r)
	})
	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		handleReset(w, r)
	})
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		handleDownload(w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
