// Package main -
package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rohit-Jung/func-redis/server"
)

func main() {

	var sigs chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	var wg sync.WaitGroup
	wg.Add(2)

	go server.RunAsyncServer(&wg)
	go server.SignalHandelling(&wg, sigs)

	wg.Wait()
}
