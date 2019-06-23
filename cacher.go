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
	cacheManager, err = cache.New(cacheProvider)
	if err != nil {
		fmt.Printf(err.Error())
	}
	if *config.Interface == "http" {
		startHTTPServer()
	} else {
		//
	}
}
