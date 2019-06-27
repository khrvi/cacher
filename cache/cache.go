package cache

import (
	"encoding/json"
	"fmt"
	"strconv"

	"./aof"
	"./cdb"
	mm "./mutex_map"
	sm "./sync_map"
)

type (
	// Cache is the basic interface expected from the backing in-memory cache
	Cache interface {
		Set(key string, value interface{}, ttl int64) error
		Get(key string) (interface{}, int64, bool, error)
		Delete(key string) error
		GetKeys() ([]string, error)
	}

	CacheManager struct {
		Provider   Cache
		CDBEnabled bool
		// prevent doubling records in AOF while restoring
		RestoreMode bool
	}

	CacheManagerError struct {
		cacheType string
	}
)

func (cme CacheManagerError) Error() string {
	return fmt.Sprintf("Cache Provider '%s' is invalid.", cme.cacheType)
}

// New returns a new resources cache.
func New(cacheType string, CDBEnabled bool, CDBPeriod int, AOFEnabled bool) (manager *CacheManager, err error) {
	if CDBEnabled {
		cdb.Init(CDBPeriod)
	}
	if AOFEnabled {
		aof.Init()
	}

	if cacheType == "mutex-map" {
		manager = &CacheManager{
			Provider: mm.New(),
		}
	} else if cacheType == "sync-map" {
		manager = &CacheManager{
			Provider: sm.New(),
		}
	} else {
		return nil, CacheManagerError{cacheType}
	}

	if CDBEnabled {
		manager.CDBEnabled = true
		restoreFromCDB(manager)
	}

	if AOFEnabled {
		restoreFromAOF(manager)
	}
	return manager, nil
}

func restoreFromCDB(cm *CacheManager) {
	cm.RestoreMode = true
	counter := 0
	iter := cdb.GetIterator()
	for iter.Next() {
		record := new(cdb.Record)
		err := json.Unmarshal([]byte(iter.Value()), &record)
		if err != nil {
			fmt.Printf("Error while unmarshaling CDB message: %s", err)
		}

		cm.Set(string(iter.Key()), record.Value, record.ExpiredAt)
		counter++
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		fmt.Printf("Error while releasing CDB iterator: %s", err)
	}
	cm.RestoreMode = false
	fmt.Printf("Restored %d records from CDB\n", counter)
}

func restoreFromAOF(cm *CacheManager) {
	cm.RestoreMode = true
	counter := 0
	from := int64(0)
	if cm.CDBEnabled {
		from = cdb.GetUpdatedAtTimestamp()
	}

	listCommands := aof.GetCommands(from)
	for _, hash := range listCommands {
		if hash["op"] == "set" {
			ttl, _ := strconv.Atoi(hash["ttl"])
			cm.Set(hash["key"], hash["value"], int64(ttl))
		} else {
			cm.Delete(hash["key"])
		}

		counter++
	}

	cm.RestoreMode = false
	fmt.Printf("Restored %d operations from AOF\n", counter)
}

func (cm *CacheManager) Get(key string) (interface{}, int64, bool, error) {
	value, expiredAt, found, err := cm.Provider.Get(key)
	if err != nil {
		// log error here
	}
	return value, expiredAt, found, err
}

func (cm *CacheManager) Set(key string, value interface{}, ttl int64) (err error) {
	if !cm.RestoreMode {
		aof.Write(key, value, ttl, "pending")
	}
	err = cm.Provider.Set(key, value, ttl)
	if !cm.RestoreMode {
		if err != nil {
			aof.Write(key, value, ttl, "failed")
		} else {
			aof.Write(key, value, ttl, "completed")
		}
	}
	if cm.CDBEnabled {
		cdb.Set(key, value, ttl)
	}
	//TODO: retry in case of error
	return err
}

func (cm *CacheManager) Delete(key string) (err error) {
	if !cm.RestoreMode {
		aof.Delete(key, "pending")
	}
	err = cm.Provider.Delete(key)
	if !cm.RestoreMode {
		if err != nil {
			aof.Delete(key, "failed")
		} else {
			aof.Delete(key, "completed")
		}
	}
	if cm.CDBEnabled {
		cdb.Delete(key)
	}
	//TODO: retry in case of error
	return err
}

func (cm *CacheManager) GetKeys() ([]string, error) {
	return cm.Provider.GetKeys()
}

func Close() {
	cdb.Close()
}
