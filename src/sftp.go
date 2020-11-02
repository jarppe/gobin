package main

import (
	"github.com/melbahja/goph"
	"log"
	"sync"
)

func sftp(wg *sync.WaitGroup, config Config, changeCh <-chan Change, exitCh <-chan bool) {
	defer wg.Done()

	client, err := goph.New(config.user, config.hostname, goph.Key(config.identityfile, ""))
	if err != nil {
		log.Fatalf("can't connect %s:%s: %s", config.user, config.hostname, err.Error())
	}
	defer client.Close()

	for {
		select {
		case change := <-changeCh:
			handleChange(client, change)

		case _ = <-exitCh:
			log.Printf("sftp: closing...")
			return
		}
	}
}

func handleChange(client *goph.Client, change Change) {
	log.Printf("handleChange: %q", change)
}
