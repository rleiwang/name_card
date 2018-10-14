// serveWs handles websocket requests from the peer.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	mutex   = &sync.Mutex{}
	logdate string
	logfile = createUpdateLog(false)
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	b *broker

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	list []NameList
}

func createUpdateLog(reset bool) string {
	logdate = time.Now().Local().Format("2006-01-02")
	fname := "update_" + logdate + ".log"
	// If the file doesn't exist, create it, or append to the file

	flags := os.O_CREATE
	if reset {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(fname, flags, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return fname
}

func appendLog(msg []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(msg); err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte{'\n'}); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func loadLog() []byte {
	mutex.Lock()
	defer mutex.Unlock()

	inFile, _ := os.Open(logfile)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	updated := []NameList{}
	for scanner.Scan() {
		var single []NameList
		err := json.Unmarshal([]byte(scanner.Text()), &single)
		if err == nil {
			updated = append(updated, single...)
		}
	}

	b, err := json.Marshal(Message{Type: "Update", List: updated})
	if err != nil {
		log.Printf("error : %v", err)
		return nil
	}
	return b
}

func (c *client) in() {
	defer func() {
		c.b.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		appendLog(message)
		var updates []NameList
		err = json.Unmarshal(message, &updates)
		if err != nil {
			log.Printf("error 1: %v", err)
			continue
		}
		outgoing, err := json.Marshal(Message{Type: "Update", List: updates})
		if err != nil {
			log.Printf("error 2: %v", err)
			continue
		}
		c.b.broadcast <- outgoing
	}
}

func (c *client) out() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(b *broker, l *NameLists, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	buf, err := json.Marshal(Message{Type: "Load", List: l.lists})
	if err != nil {
		log.Println(err)
		return
	}
	cli := &client{b: b, conn: conn, send: make(chan []byte, 256), list: l.lists}
	b.register <- cli

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go cli.out()
	go cli.in()

	if len(l.lists) > 0 {
		cli.send <- buf
		cli.send <- loadLog()
	}
}

func handleUpload(l *NameLists, w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer file.Close()

	l.lists = readFiles(file)
	w.Header().Add("Access-Control-Allow-Origin", "*")
	logfile = createUpdateLog(true)
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	logfile = createUpdateLog(true)
	w.Header().Add("Access-Control-Allow-Origin", "*")
}

func readFiles(f io.Reader) []NameList {
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	of, err := os.OpenFile("attendance.csv", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer of.Close()

	names := []NameList{}
	// ID,Family ID,Chinese Name,Last Name,First Name
	scanner.Scan()
	for scanner.Scan() {
		fmt.Fprintln(of, scanner.Text())
		names = append(names, *parseToNameList(scanner.Text()))
	}

	sort.Slice(names, func(i, j int) bool { return names[i].FID > names[j].FID })
	for i := 0; i < len(names); i++ {
		names[i].ID = i
	}
	return names
}

func parseToNameList(t string) *NameList {
	s := strings.Split(t, ",")
	uid, _ := strconv.Atoi(s[0])
	fid, _ := strconv.Atoi(s[1])
	return &NameList{UID: uid, FID: fid, ChineseName: s[2], Family: s[3], First: s[4], Absent: true}
}

func loadFileIfExists() []NameList {
	names := []NameList{}
	if _, err := os.Stat("attendance.csv"); err != nil && os.IsNotExist(err) {
		return names
	}

	f, err := os.OpenFile("attendance.csv", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	scanner.Scan()
	for scanner.Scan() {
		names = append(names, *parseToNameList(scanner.Text()))
	}
	sort.Slice(names, func(i, j int) bool { return names[i].FID > names[j].FID })
	for i := 0; i < len(names); i++ {
		names[i].ID = i
	}
	return names
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	modtime := time.Now()
	content := bytes.NewReader(downloadLog())

	// ServeContent uses the name for mime detection
	name := "report_" + logdate + ".csv"

	// tell the browser the returned content should be downloaded
	w.Header().Add("Content-Disposition", "attachment;filename="+name)
	w.Header().Add("Content-Type", "application/text")

	http.ServeContent(w, r, name, modtime, content)
}

func downloadLog() []byte {
	mutex.Lock()
	defer mutex.Unlock()

	if logfile == "" {
		return []byte{}
	} else if _, err := os.Stat(logfile); err != nil && os.IsNotExist(err) {
		return []byte{}
	}
	inFile, _ := os.Open(logfile)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	attendees := map[int]NameList{}
	for scanner.Scan() {
		var single []NameList
		err := json.Unmarshal([]byte(scanner.Text()), &single)
		if err == nil {
			for _, nl := range single {
				if nl.Absent {
					delete(attendees, nl.ID)
				} else {
					attendees[nl.ID] = nl
				}
			}
		}
	}
	b := []byte{}
	for _, v := range attendees {
		b = append(b, []byte(fmt.Sprintf("%d,%s,%s,%s\n", v.UID, v.ChineseName, v.First, v.Family))...)
	}

	return b
}
