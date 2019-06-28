package main

import (
	l "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"./cache"
	"./config"
	"github.com/google/logger"
)

var (
	cacheManager cache.Cache
	logfile      *os.File
	log          *l.Logger
)

const (
	logFolder = "./log/"
	logName   = "cacher.log"
)

func main() {
	prepareLogger()
	cacheProvider := *config.CacheType
	var err error
	cacheManager, err = cache.New(cacheProvider, log, *config.CDBEnabled, *config.CDBPeriod, *config.AOFEnabled)
	if err != nil {
		log.Fatalf("Error while initializing cache manager: %s", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range signals {
			log.Println("Shutting down Cacher...")
			cache.Close()
			logfile.Close()
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

func prepareLogger() {
	os.MkdirAll(logFolder, os.ModePerm)
	logPath := logFolder + logName
	var err error
	if !fileExists(logPath) {
		logfile, err = os.Create(logPath)
	} else {
		logfile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	}
	if err != nil {
		logger.Fatalf("Failed to create/open log file: %v", err)
	}
	log = l.New(logfile, "Cacher: ", l.LstdFlags)
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
