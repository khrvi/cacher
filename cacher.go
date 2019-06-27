package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"./cache"
	"./config"
)

var cacheManager cache.Cache

func main() {
	var err error
	cacheProvider := *config.CacheType
	cacheManager, err = cache.New(cacheProvider, *config.CDBEnabled, *config.CDBPeriod, *config.AOFEnabled)
	if err != nil {
		fmt.Printf(err.Error())
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range signals {
			fmt.Printf("Shutting down Cacher...")
			cache.Close()
			time.Sleep(2 * time.Second)
			os.Exit(1)
		}
	}()

	if *config.Interface == "http" {
		startHTTPServer()
	} else {
		startTelNetServer()
	}
}
