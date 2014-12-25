package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Command byte

const (
	Start Command = iota
	Stop
	Note
	Show
)

type TrackDB map[string][]TrackEntry

type TrackEntry struct {
	Start string
	Stop  string
	Notes []string
}

func (c Command) String() string {
	switch c {
	case Start:
		return "start"
	case Stop:
		return "stop"
	case Note:
		return "note"
	case Show:
		return "show"
	default:
		return "unknown"
	}
}

func main() {
	command, tag, note, err := parseArgs(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbFile := path.Join(usr.HomeDir, ".track.toml")

	db, err := readDB(dbFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch command {
	case Start:
		doStart(tag, db)
	case Stop:
		doStop(tag, db)
	case Note:
		doNote(tag, note, db)
	case Show:
		doShow(tag, db)
	}

	if command != Show {
		if err := writeDB(dbFile, db); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func doStart(tag string, db TrackDB) {
	entries := db[tag]
	var lastEntry TrackEntry
	if len(entries) > 0 {
		lastEntry = entries[len(entries)-1]
	} else {
		lastEntry = TrackEntry{}
	}
	if lastEntry.Start != "" && lastEntry.Stop == "" {
		fmt.Printf("Already started at %s\n", lastEntry.Start)
		return
	}
	db[tag] = append(entries, TrackEntry{Start: time.Now().Format("2006-01-02 15:04:05")})
}

func doStop(tag string, db TrackDB) {
	entries := db[tag]
	if len(entries) == 0 {
		fmt.Printf("Unknown tag \"%s\"\n", tag)
		return
	}
	lastEntry := &entries[len(entries)-1]
	if lastEntry.Stop != "" {
		fmt.Printf("Already stopped at %s\n", lastEntry.Stop)
		return
	}
	lastEntry.Stop = time.Now().Format("2006-01-02 15:04:05")
}

func doNote(tag string, note string, db TrackDB) {
	entries := db[tag]
	if len(entries) == 0 {
		fmt.Printf("Unknown tag \"%s\"\n", tag)
		return
	}
	lastEntry := &entries[len(entries)-1]
	if lastEntry.Stop != "" {
		fmt.Println("Not currently tracking anything")
		return
	}
	lastEntry.Notes = append(lastEntry.Notes, note)
}

func doShow(tag string, db TrackDB) {
	if tag == "all" {
		for t, _ := range db {
			doShow(t, db)
		}
		return
	}

	fmt.Println(tag)
	for _, entry := range db[tag] {
		fmt.Printf("\t%s - %s\n", entry.Start, entry.Stop)
		for _, note := range entry.Notes {
			fmt.Printf("\t\t%s\n", note)
		}
	}
}

func readDB(dbFile string) (TrackDB, error) {
	var db TrackDB
	if _, err := toml.DecodeFile(dbFile, &db); err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Create(dbFile); err != nil {
				return nil, err
			}
			return db, nil
		}
		return nil, err
	}
	return db, nil
}

func writeDB(dbFile string, db TrackDB) error {
	w, err := os.Create(dbFile)
	if err != nil {
		return err
	}
	encoder := toml.NewEncoder(w)
	return encoder.Encode(db)
}

// track [<command>] [<project>] [<note>]
func parseArgs(args []string) (command Command, tag string, note string, err error) {
	// TODO Print help and exit
	if len(args) == 1 {
		command = Show
		tag = "all"
		return
	}

	switch args[1] {
	case "start":
		command = Start
	case "stop":
		command = Stop
	case "show":
		command = Show
	case "note":
		command = Note
		if len(args) < 4 {
			err = errors.New(`Error: Missing note`)
		} else {
			note = strings.Join(args[3:], " ")
		}
	default:
		err = errors.New(`Error: Invalid command (expected one of "start", "stop", "show", "note")`)
		return
	}

	if len(args) < 3 && command != Show {
		err = errors.New(`Error: Missing tag`)
	} else if len(args) < 3 {
		tag = "all"
	} else {
		tag = args[2]
	}

	return
}
