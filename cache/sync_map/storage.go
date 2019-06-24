package sync_map

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
		values *sync.Map
	}
)

func New() *Storage {
	return &Storage{
		values: &sync.Map{},
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
	s.values.Store(key, record)
	return nil
}

func (s *Storage) Get(key string) (interface{}, int64, bool, error) {
	raw, found := s.values.Load(key)
	if !found {
		return nil, 0, false, nil
	}
	record := raw.(Record)
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
	s.values.Delete(key)
	return nil
}
