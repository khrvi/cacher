package mutex_map

import (
	"encoding/json"
	"sync"
	"time"
)

type (
	Record struct {
		Value     []byte
		ExpiredAt int64
	}

	Storage struct {
		mu     sync.RWMutex
		values map[string]Record
	}
)

func New() *Storage {
	return &Storage{
		values: make(map[string]Record),
	}
}

func (s *Storage) Set(key string, value interface{}, ttl int64) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	record := Record{Value: data}
	if ttl != 0 {
		record.ExpiredAt = time.Now().Add(time.Second * time.Duration(ttl)).Unix()
	}

	s.mu.Lock()
	s.values[key] = record
	s.mu.Unlock()
	return nil
}

func (s *Storage) Get(key string) (interface{}, int64, bool, error) {
	s.mu.RLock()
	record, found := s.values[key]
	s.mu.RUnlock()
	if !found {
		return nil, 0, false, nil
	}
	var data interface{}
	err := json.Unmarshal(record.Value, &data)
	if err != nil {
		return nil, 0, false, err
	}
	// expire record if time has come
	if record.ExpiredAt > 0 && (time.Now().Unix() >= record.ExpiredAt) {
		return nil, 0, false, nil
	}

	return data, record.ExpiredAt, true, nil
}

func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	delete(s.values, key)
	s.mu.Unlock()
	return nil
}

func (s *Storage) GetKeys() ([]string, error) {
	keys := make([]string, 0)
	s.mu.RLock()
	for key := range s.values {
		keys = append(keys, key)
	}
	s.mu.RUnlock()
	return keys, nil
}
