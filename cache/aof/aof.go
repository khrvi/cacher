package aof

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var path = "./data/aof/aof.log"

func Init() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     30, //days
	})
}

func Infof(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Write(key string, value interface{}, ttl int64, state string) {
	log.Printf(" set %s %s %d - %s", key, string(marshal(value)), ttl, state)
}

func Delete(key string, state string) {
	log.Printf(" delete %s - %s", key, state)
}

func marshal(value interface{}) []byte {
	data, _ := json.Marshal(value)
	return data
}

// GetCommands parses AOF log and prepare list of all commands that were requested from passed timestamp
func GetCommands(from int64) (result []map[string]string) {
	if from == 0 {
		fmt.Println("AOF: restoring all records...")
	} else {
		fmt.Printf("AOF: restoring from %d timestamp...\n", from)
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return result
	}
	defer file.Close()

	timezone, _ := time.Now().Local().Zone()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		dateTime, _ := time.Parse("2006/01/02 15:04:05 MST", parts[0]+" "+parts[1]+" "+timezone)
		timestamp := dateTime.UTC().Unix()
		if from != 0 {
			if timestamp >= from {
				record := prepareRecord(parts)
				if record != nil {
					result = append(result, record)
				}
			}
		} else {
			result = append(result, prepareRecord(parts))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return result
}

func prepareRecord(parts []string) (record map[string]string) {
	var status string
	if parts[3] == "delete" {
		record = map[string]string{
			"op":  parts[3],
			"key": parts[4],
		}
		status = parts[6]
	} else {
		record = map[string]string{
			"op":    parts[3],
			"key":   parts[4],
			"value": parts[5],
			"ttl":   parts[6],
		}
		status = parts[8]
	}

	// consider only pending commands
	if status == "pending" {
		return record
	} else {
		return nil
	}
}
