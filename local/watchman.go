package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type Change struct {
	Version         string
	Clock           string
	Files           []File
	Root            string
	Subscription    string
	Unilateral      bool
}

type File struct {
	Mode   int
	Exists bool
	Size   int
	Type   string
	Name   string
}

func watchman(wg *sync.WaitGroup, config Config, changeCh chan<- Change, exitCh <-chan bool) {
	defer wg.Done()

	getSockNameCmd := exec.Command("watchman", "get-sockname")
	out, err := getSockNameCmd.Output()
	if err != nil {
		log.Fatalf("can't get watchman socket name, have you installed Watchman? (https://facebook.github.io/watchman/)")
	}
	var socketName struct {
		Version  string
		Sockname string
	}
	parse(out, &socketName)

	watchman := MakeWatchman(socketName.Sockname)
	defer watchman.Close()

	watchman.subscribe(config.source)
	watchman.listenChanges(changeCh)

	log.Printf("gobin: Ready")
	<-exitCh
}

type Watchman struct {
	socket net.Conn

	writer  *bufio.Writer
	encoder *json.Encoder

	reader  *bufio.Reader
	scanner *bufio.Scanner

	subscription string
}

func MakeWatchman(socketName string) *Watchman {
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
		"gobin:" + strconv.Itoa(os.Getpid()),
	}
}

func (watchman *Watchman) Close() {
	watchman.socket.Close()
}

func (watchman *Watchman) subscribe(watchRoot string) {
	req := []interface{}{
		"subscribe",
		watchRoot,
		watchman.subscription,
		map[string]interface{}{
			"expression": []interface{}{
				"allof",
				[]interface{}{"match", "**/*", "wholename", map[string]interface{}{"includedotfiles": true}},
				[]interface{}{"not", []interface{}{"dirname", ".git"}},
				[]interface{}{"not", []interface{}{"match", ".git", "wholename"}},
				[]interface{}{"not", []interface{}{"dirname", ".idea"}},
				[]interface{}{"not", []interface{}{"match", ".idea", "wholename"}},
			},
			"fields": []interface{}{"name", "type", "mode", "size", "exists"},
		},
	}

	var resp struct {
		Version   string `json:"version"`
		Subscribe string `json:"subscribe"`
	}
	if !watchman.sendCommand(req, &resp) {
		log.Fatalf("socket closed at subscribe")
	}
}

func (watchman *Watchman) sendCommand(req interface{}, resp interface{}) bool {
	watchman.encoder.Encode(req)
	watchman.writer.Flush()
	return watchman.read(resp)
}

func (watchman *Watchman) read(resp interface{}) bool {
	ok := watchman.scanner.Scan()
	if !ok {
		return false // EOF
	}
	err := json.Unmarshal(watchman.scanner.Bytes(), resp)
	if err != nil {
		log.Fatalf("can't parse watch-project command response: %q", watchman.scanner.Text())
	}
	return true
}

func (watchman *Watchman) listenChanges(changesCh chan<- Change) {
	go func() {
		defer close(changesCh)
		for {
			change := Change{}
			ok := watchman.read(&change)
			if !ok {
				return
			}
			changesCh <- change
		}
	}()
}

func parse(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Fatalf("can't parse: %v (%T)", err, err)
	}
}

