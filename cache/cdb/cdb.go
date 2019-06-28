package cdb

import (
	"encoding/json"
	"log"
	l "log"
	"time"

	"./leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type Record struct {
	Value     interface{}
	ExpiredAt int64
}

var (
	directWrite bool
	// used this odd key for storing timestamp at the same db, or create a new one
	updatedAtTimestampKey = "--updated_at_timestamp--"
)

func Init(period int, log *l.Logger) {
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
			refreshUpdatedAtTimestamp()
			err := leveldb.SaveBatch()
			if err != nil {
				log.Fatalf("Error while saving batch: %s", err)
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
		refreshUpdatedAtTimestamp()
		err = leveldb.WriteKey(key, data)
	} else {
		err = leveldb.AddToBatch(key, data)
	}
	if err != nil {
		log.Fatalf("Can't save message to CDB. %s", err)
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
		refreshUpdatedAtTimestamp()
		err = leveldb.DelKey(data)
	} else {
		err = leveldb.RemoveFromBatch(data)
	}

	if err != nil {
		log.Fatalf("Error while cleaning CDB key '%s': %s", key, err)
	}
	return nil
}

func refreshUpdatedAtTimestamp() {
	Set(updatedAtTimestampKey, time.Now().Unix(), 0)
}

func GetUpdatedAtTimestamp() int64 {
	value, err := leveldb.ReadKey(updatedAtTimestampKey)
	if err != nil {
		if !leveldb.IsNotFound(err) {
			log.Println("UpdatedAtTimestamp is not found.")
		} else {
			log.Fatalf("Error while getting updatedAtTimestampKey from CDB: %s", err)
		}

		return 0
	} else {
		record := new(Record)
		err := json.Unmarshal([]byte(value), &record)
		if err != nil {
			log.Fatalf("Error while unmarshaling CDB message: %s", err)
		}

		return int64(record.Value.(float64))
	}
}

func GetIterator() iterator.Iterator {
	return leveldb.Iterator()
}

func Close() {
	if directWrite == false {
		refreshUpdatedAtTimestamp()
		// save existing batch before exist
		err := leveldb.SaveBatch()
		if err != nil {
			log.Fatalf("Error while saving batch: %s", err)
		}

	}
	leveldb.Close()
}
