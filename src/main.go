package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

//
// Talk to WatchMan
// sync
//

func main() {
	log.SetPrefix("gobin: ")

	config := loadConfig()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitCh := make(chan bool)
	changeCh := make(chan Change)

	go func() {
		defer close(exitCh)
		<-sigCh
		signal.Reset(
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)
		log.Printf("Closing...")
	}()

	var wg sync.WaitGroup
	go sftp(&wg, config, changeCh)
	go watchman(&wg, config, changeCh, exitCh)
	wg.Add(2)

	log.Printf("Ready")
	wg.Wait()
	log.Printf("Terminating")
	os.Exit(0)
}
