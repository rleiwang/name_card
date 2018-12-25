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
	"path"
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
	logfname := path.Join(*dir, "update_"+logdate+".log")
	// If the file doesn't exist, create it, or append to the file

	flags := os.O_CREATE
	if reset {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(logfname, flags, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return logfname
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
	of, err := os.OpenFile(*fname, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer of.Close()

	return parseToNameLists(f, of)
}

func parseToNameLists(r io.Reader, of *os.File) []NameList {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	names := []NameList{}

	// Family Name,Family ID,ID,Chinese Name,Last Name,First Name
	scanner.Scan()
	if of != nil {
		fmt.Fprintln(of, scanner.Text())
	}

	for scanner.Scan() {
		t := scanner.Text()
		if of != nil {
			fmt.Fprintln(of, t)
		}
		s := strings.Split(t, ",")
		fid, _ := strconv.Atoi(s[1])
		uid, _ := strconv.Atoi(s[2])
		names = append(names, NameList{UID: uid, FID: fid, ChineseName: s[3], Family: s[0], Last: s[4], First: s[5], Absent: true})
	}
	sort.Slice(names, func(i, j int) bool {
		v := strings.Compare(names[i].Family, names[j].Family)
		if v > 0 {
			return false
		} else if v < 0 {
			return true
		}
		v = names[i].FID - names[j].FID
		if v > 0 {
			return true
		} else if v < 0 {
			return false
		}
		return names[i].UID < names[j].UID
	})
	// jsx uses array index as ID
	for i := 0; i < len(names); i++ {
		names[i].IDX = i
	}
	return names
}

func loadFileIfExists() []NameList {
	if _, err := os.Stat(*fname); err != nil && os.IsNotExist(err) {
		return []NameList{}
	}

	f, err := os.OpenFile(*fname, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	return parseToNameLists(f, nil)
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
					delete(attendees, nl.UID)
				} else {
					attendees[nl.UID] = nl
				}
			}
		}
	}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	for _, v := range attendees {
		w.WriteString(fmt.Sprintln(v.UID))
	}
	w.Flush()

	return buf.Bytes()
}
