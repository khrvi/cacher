package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"./config"
	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
	"github.com/reiver/go-telnet/telsh"
)

type Result struct {
	Status       string      `json:"status"`
	Value        interface{} `json:"value,omitempty"`
	ExpiredAt    string      `json:"expired_at,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

func startTelNetServer() {
	shellHandler := telsh.NewShellHandler()
	shellHandler.WelcomeMessage = `
  /$$$$$$                      /$$                          
 /$$__  $$                    | $$                          
| $$  \__/  /$$$$$$   /$$$$$$$| $$$$$$$   /$$$$$$   /$$$$$$ 
| $$       |____  $$ /$$_____/| $$__  $$ /$$__  $$ /$$__  $$
| $$        /$$$$$$$| $$      | $$  \ $$| $$$$$$$$| $$  \__/
| $$    $$ /$$__  $$| $$      | $$  | $$| $$_____/| $$      
|  $$$$$$/|  $$$$$$$|  $$$$$$$| $$  | $$|  $$$$$$$| $$      
\______/  \_______/ \_______/|__/  |__/ \_______/|__/      
`

	commandName := "get"
	commandProducer := telsh.ProducerFunc(getValuePruducer)
	shellHandler.Register(commandName, commandProducer)
	commandName = "set"
	commandProducer = telsh.ProducerFunc(setValuePruducer)
	shellHandler.Register(commandName, commandProducer)
	commandName = "delete"
	commandProducer = telsh.ProducerFunc(deleteValuePruducer)
	shellHandler.Register(commandName, commandProducer)

	address := fmt.Sprintf("%s:%s", *config.ServerIP, *config.ServerPort)
	if err := telnet.ListenAndServe(address, shellHandler); nil != err {
		panic(err)
	}
}

func getValueHandler(stdin io.ReadCloser, stdout io.WriteCloser, stderr io.WriteCloser, args ...string) error {
	if len(args) == 1 {
		key := args[0]
		value, expiredAt, found, err := cacheManager.Get(key)
		if err != nil {
			errorMessage := fmt.Sprintf("Error occured while Get value '%s' from cache.\n\r", key)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}
		if !found {
			errorMessage := fmt.Sprintf("Key '%s' not found.\n\r", key)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}

		result := Result{Status: "ok"}
		if value != "" {
			result.Value = value
			if expiredAt != 0 {
				result.ExpiredAt = time.Unix(expiredAt, 0).Format("2006-01-02 15:04:05")
			}

		}

		b, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err)
			errorMessage := fmt.Sprintf("Error occured while Get value '%s' from cache.\n\r", key)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}
		oi.LongWriteString(stdout, string(b)+"\n\r")

	} else {
		oi.LongWriteString(stdout, "Command GET requires one parameter: 'Key'.")
	}

	return nil
}

func getValuePruducer(ctx telnet.Context, name string, args ...string) telsh.Handler {
	return telsh.PromoteHandlerFunc(getValueHandler, args...)
}

func setValueHandler(stdin io.ReadCloser, stdout io.WriteCloser, stderr io.WriteCloser, args ...string) error {
	if len(args) == 2 || len(args) == 3 {
		key := args[0]
		value := args[1]
		var ttl int64
		if len(args) == 3 {
			var err error
			ttl, err = strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				oi.LongWriteString(stdout, "TTL value is invalid!\n\r")
				return nil
			}
		}

		var rawValue interface{}
		err := json.Unmarshal([]byte(value), &rawValue)
		if err != nil {
			errorMessage := fmt.Sprintf("Error occured while adding new key/value pair: %s - %s\n\r", key, value)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}
		error := cacheManager.Set(key, rawValue, ttl)
		if error != nil {
			fmt.Println(err)
			errorMessage := fmt.Sprintf("Error occured while adding new key/value pair: %s - %s\n\r", key, value)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}

		result := &Result{Status: "Ok"}
		b, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err)
			errorMessage := fmt.Sprintf("Error occured while adding new key/value pair: %s - %s\n\r", key, value)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}
		oi.LongWriteString(stdout, string(b)+"\n\r")

	} else {
		oi.LongWriteString(stdout, "Command Set requires two params: 'Key' and 'Value'. Optional param is 'TTL'.")
	}
	return nil
}

func setValuePruducer(ctx telnet.Context, name string, args ...string) telsh.Handler {
	return telsh.PromoteHandlerFunc(setValueHandler, args...)
}

func deleteValueHandler(stdin io.ReadCloser, stdout io.WriteCloser, stderr io.WriteCloser, args ...string) error {
	if len(args) == 1 {
		//TODO: validate that key exists
		key := args[0]
		err := cacheManager.Delete(key)
		if err != nil {
			errorMessage := fmt.Sprintf("Error occured while deleting key: %s", key)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}

		result := &Result{Status: "Ok"}
		b, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err)
			errorMessage := fmt.Sprintf("Error occured while deleting key: %s\n\r", key)
			oi.LongWriteString(stdout, errorMessage)
			return nil
		}
		oi.LongWriteString(stdout, string(b)+"\n\r")

	} else {
		oi.LongWriteString(stdout, "Command DELETE requires one parameter: 'Key'.")
	}

	return nil
}

func deleteValuePruducer(ctx telnet.Context, name string, args ...string) telsh.Handler {
	return telsh.PromoteHandlerFunc(deleteValueHandler, args...)
}
