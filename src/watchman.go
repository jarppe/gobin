package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"sync"
)

func watchman(wg *sync.WaitGroup, config Config, changeCh chan<- Change, exitCh <-chan bool) {
	defer wg.Done()

	getSockNameCmd := exec.Command("watchman", "get-sockname")
	out, err := getSockNameCmd.Output()
	if err != nil {
		log.Fatalf("can't get watchman socket name, have you installed Watchman? (https://facebook.github.io/watchman/)")
	}

	type SocketName struct {
		Version  string
		Sockname string
	}

	var socketName SocketName
	err = json.Unmarshal(out, &socketName)
	if err != nil {
		log.Fatalf("can't parse watchman socket response")
	}

	watchman := MakeWatchman(socketName.Sockname, changeCh)

	watchProjectResp := watchman.watchProject(config.source)

	log.Printf("watching successfully, watch = %q relp = %q", watchProjectResp.Watch, watchProjectResp.RelativePath)

	<-exitCh
	log.Printf("watchman closing...")
}

func (watchman *Watchman) watchProject(dirName string) *WatchProjectResp {
	watchProjectResp := &WatchProjectResp{}
	watchman.sendCommand([]string{"watch-project", dirName}, watchProjectResp)
	return watchProjectResp
}

func (watchman *Watchman) sendCommand(req interface{}, resp interface{}) {
	watchman.encoder.Encode(req)
	watchman.writer.Flush()

	watchman.scanner.Scan()
	err := json.Unmarshal([]byte(watchman.scanner.Text()), resp)
	if err != nil {
		log.Fatalf("can't parse watch-project command response: %q", watchman.scanner.Text())
	}
}


func MakeWatchman(socketName string, changeCh chan<- Change) *Watchman {
	socket, err := net.Dial("unix", socketName)
	if err != nil {
		log.Fatalf("can't open Watchman socket %s: %s", socketName, err.Error())
	}

	writer := bufio.NewWriter(socket)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	reader := bufio.NewReader(socket)
	scanner := bufio.NewScanner(reader)

	return &Watchman{
		socket,
		writer,
		encoder,
		reader,
		scanner,
		changeCh,
	}
}

type Watchman struct {
	socket net.Conn

	writer  *bufio.Writer
	encoder *json.Encoder

	reader  *bufio.Reader
	scanner *bufio.Scanner

	changesCh chan<- Change
}

type Change struct {
	Version      string   `json:"version"`
	Clock        string   `json:"clock"`
	Files        []string `json:"files"`
	Root         string   `json:"root"`
	Subscription string   `json:"subscription"`
}

type WatchProjectResp struct {
	Version      string
	Watch        string
	RelativePath string `json: relative_path`
}


//   const subscribe = ["subscribe", watch, subscriptionName, {
//    expression: ["allof",
//      ["match", "**/*", "wholename", {"includedotfiles": true}],
//      // TODO: SHould have .robinignore
//      ["not", ["dirname", ".git"]],
//      ["not", ["match", ".git", "wholename"]],
//      ["not", ["dirname", ".idea"]],
//      ["not", ["match", ".idea", "wholename"]],
//      // TODO: Should be configurable:
//      ["not", ["dirname", "node_modules"]],
//      ["not", ["match", "node_modules", "wholename"]],
//      ["not", ["dirname", "dist"]],
//      ["not", ["match", "dist", "wholename"]],
//    ],
//    fields: ["name", "type", "mode", "size", "exists"],
//    relative_root: relativePath,
//  }]

// {
//  "version":   "1.6",
//  "subscribe": "mysubscriptionname"
// }
