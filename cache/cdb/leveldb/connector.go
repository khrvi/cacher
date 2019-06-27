package leveldb

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

var dbi *leveldb.DB
var batch *leveldb.Batch

func InitConnection() {
	var err error
	dbi, err = leveldb.OpenFile("./data/cdb", nil)
	if err != nil {
		log.Fatalln(err)
	}

	batch = new(leveldb.Batch)

	counter := 0
	iter := Iterator()
	for iter.Next() {
		counter += 1
	}
	log.Printf("LevelDB contains %d records...", counter)
}

func ReadKey(key string) (val string, err error) {
	data, err := dbi.Get([]byte(key), nil)
	if err != nil {
		return "", err
	}
	return string(data), err
}

func WriteKey(key string, value []byte) (err error) {
	err = dbi.Put([]byte(key), value, nil)
	if err != nil {
		return err
	}
	return nil
}

func AddToBatch(key string, value []byte) (err error) {
	batch.Put([]byte(key), value)
	return nil
}

func DelKey(key []byte) (err error) {
	err = dbi.Delete(key, nil)
	if err != nil {
		return err
	}
	return nil
}

func RemoveFromBatch(key []byte) (err error) {
	batch.Delete(key)
	return nil
}

func IsNotFound(err error) bool {
	return err == leveldb.ErrNotFound
}

func Iterator() iterator.Iterator {
	return dbi.NewIterator(nil, nil)
}

func SaveBatch() (err error) {
	err = dbi.Write(batch, nil)
	if err != nil {
		return err
	}
	// recreate a new batch
	batch = new(leveldb.Batch)
	return nil
}

func Close() {
	if batch != nil {
		// save existing batch before exist
		SaveBatch()
	}
	if dbi != nil {
		dbi.Close()
	}
}
