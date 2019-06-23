package main

import (
	"fmt"

	"./cache"
)

var cacheManager cache.Cache

func main() {
	var err error
	cacheProvider := "mutex-map"
	cacheManager, err = cache.New(cacheProvider)
	if err != nil {
		fmt.Printf(err.Error())
	}
	startHTTPServer()
}
