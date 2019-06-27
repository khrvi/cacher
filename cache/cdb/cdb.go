package cdb

import (
	"encoding/json"
	"fmt"
	"time"

	"./leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type Record struct {
	Value     interface{}
	ExpiredAt int64
}

var directWrite bool

func Init(period int) {
	leveldb.InitConnection()

	if period > 0 {
		directWrite = false
		go func() {
			initPeriodicBackup(period)
		}()
	} else {
		directWrite = true
	}
}

func initPeriodicBackup(period int) {
	backupTicker := time.NewTicker(time.Second * time.Duration(period))
	for {
		select {
		case <-backupTicker.C:
			fmt.Println("Saving batch of operations...")
			err := leveldb.SaveBatch()
			if err != nil {
				fmt.Printf("Error while saving batch: %s", err)
			}
		}
	}

}

func Set(key string, value interface{}, ttl int64) (err error) {
	record := Record{Value: value}
	if ttl != 0 {
		record.ExpiredAt = time.Now().Add(time.Second * time.Duration(ttl)).Unix()
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	if directWrite == true {
		err = leveldb.WriteKey(key, data)
	} else {
		err = leveldb.AddToBatch(key, data)
	}
	if err != nil {
		fmt.Printf("Can't save message to leveldb. %s", err)
		return err
	}
	return nil
}

func Delete(key string) (err error) {
	data, err := json.Marshal(key)
	if err != nil {
		return err
	}
	if directWrite == true {
		err = leveldb.DelKey(data)
	} else {
		err = leveldb.RemoveFromBatch(data)
	}

	if err != nil {
		fmt.Printf("Error while cleaning leveldb key '%s': %s", key, err)
	}
	return nil
}

func GetIterator() iterator.Iterator {
	return leveldb.Iterator()
}

func Close() {
	leveldb.Close()
}
