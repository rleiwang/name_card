package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var input = flag.String("input", "", "input file")
var output = flag.String("out", "", "output file")

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

func main() {
	flag.Parse()

	if *input == "" || *output == "" {
		flag.Usage()
		return
	}

	t, err := time.Parse("update_2006-01-02.log", filepath.Base(*input))
	if err != nil {
		fmt.Printf("log file must be in format: update_2006-01-02.log")
		panic(err)
	}
	fmt.Printf("processing date: %v \n", t)

	inFile, err := os.Open(*input)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	outFile, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

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

	recdate := t.Format("01/02/2006")
	for _, v := range attendees {
		outFile.WriteString(fmt.Sprintln(v.UID, ",", recdate))
	}
}
