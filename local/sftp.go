package main

import (
	"github.com/melbahja/goph"
	"log"
	"sync"
)

func sftp(wg *sync.WaitGroup, config Config, changeCh <-chan Change) {
	defer wg.Done()

	client, err := goph.New(config.user, config.hostname, goph.Key(config.identityfile, ""))
	if err != nil {
		log.Fatalf("can't connect %s:%s: %s", config.user, config.hostname, err.Error())
	}

	for change := range changeCh {
		handleChange(client, change)
	}

	log.Printf("sftp: closing...")

	err = client.Close()
	if err != nil {
		log.Printf("sftp: error while closing SSH client: %s", err.Error())
	}
}

func handleChange(client *goph.Client, change Change) {
	log.Printf("handleChange:")
	for _, file := range change.Files {
		log.Printf("   [%s] (%s)", file.Name, file.Type)
	}
}
