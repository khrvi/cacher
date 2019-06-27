package main

import (
	"fmt"

	"./cache"
	"./config"
)

var cacheManager cache.Cache

func main() {
	var err error
	cacheProvider := *config.CacheType
	cacheManager, err = cache.New(cacheProvider, *config.CDBEnabled, *config.CDBPeriod)
	if err != nil {
		fmt.Printf(err.Error())
	}

	defer func() {
		cache.Close()
	}()

	if *config.Interface == "http" {
		startHTTPServer()
	} else {
		startTelNetServer()
	}
}
