package cache

import (
	"encoding/json"
	"fmt"

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
	}

	CacheManagerError struct {
		cacheType string
	}
)

func (cme CacheManagerError) Error() string {
	return fmt.Sprintf("Cache Provider '%s' is invalid.", cme.cacheType)
}

// New returns a new resources cache.
func New(cacheType string, CDBEnabled bool, CDBPeriod int) (manager *CacheManager, err error) {
	if CDBEnabled {
		cdb.Init(CDBPeriod)
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
		RestoreTo(manager)
	}
	return manager, nil
}

func RestoreTo(cm *CacheManager) {
	iter := cdb.GetIterator()
	for iter.Next() {
		record := new(cdb.Record)
		err := json.Unmarshal([]byte(iter.Value()), &record)
		if err != nil {
			fmt.Printf("Error while unmarshaling leveldb message: %s", err)
		}

		cm.Set(string(iter.Key()), record.Value, record.ExpiredAt)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		fmt.Printf("Error while releasing leveldb iterator: %s", err)
	}
}

func (cm *CacheManager) Get(key string) (interface{}, int64, bool, error) {
	value, expiredAt, found, err := cm.Provider.Get(key)
	if err != nil {
		// log error here
	}
	return value, expiredAt, found, err
}

func (cm *CacheManager) Set(key string, value interface{}, ttl int64) (err error) {
	err = cm.Provider.Set(key, value, ttl)
	if cm.CDBEnabled {
		cdb.Set(key, value, ttl)
	}
	//TODO: retry in case of error
	return err
}

func (cm *CacheManager) Delete(key string) (err error) {
	err = cm.Provider.Delete(key)
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
