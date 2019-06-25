package cache

import (
	"fmt"

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
		Provider Cache
	}

	CacheManagerError struct {
		cacheType string
	}
)

func (cme CacheManagerError) Error() string {
	return fmt.Sprintf("Cache Provider '%s' is invalid.", cme.cacheType)
}

// New returns a new resources cache.
func New(cacheType string) (*CacheManager, error) {
	if cacheType == "mutex-map" {
		return &CacheManager{
			Provider: mm.New(),
		}, nil
	} else if cacheType == "sync-map" {
		return &CacheManager{
			Provider: sm.New(),
		}, nil
	}

	return nil, CacheManagerError{cacheType}
}

func (cm *CacheManager) Get(key string) (interface{}, int64, bool, error) {
	value, expiredAt, found, err := cm.Provider.Get(key)
	if err != nil {
		// log error here
	}
	return value, expiredAt, found, err
}

func (cm *CacheManager) Set(key string, value interface{}, ttl int64) error {
	return cm.Provider.Set(key, value, ttl)
}

func (cm *CacheManager) Delete(key string) error {
	return cm.Provider.Delete(key)
}

func (cm *CacheManager) GetKeys() ([]string, error) {
	return cm.Provider.GetKeys()
}
