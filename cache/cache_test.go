package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	TestRecord struct {
		Value string
	}
)

func TestNew(t *testing.T) {
	for _, name := range []string{"sync-map", "mutex-map"} {
		_, err := New(name)
		if err != nil {
			t.Fatalf("Provider '%s' failed to init: %v", name, err)
		}
	}

	provider, err := New("wrong_provider")
	assert.Nil(t, provider)
	assert.Equal(t, "Cache Provider 'wrong_provider' is invalid.", err.Error())
}

func TestGet(t *testing.T) {
	provider, _ := New("sync-map")
	// test missed key
	value, expiredAt, found, err := provider.Get("test")
	assert.Nil(t, value)
	assert.Equal(t, int64(0), expiredAt)
	assert.False(t, found)
	assert.Nil(t, err)

	// prepare key/value pair
	err = provider.Set("test", "value", 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set: %v", err)
	}
	// return existing value
	value, expiredAt, found, err = provider.Get("test")
	assert.Equal(t, "value", value)
	assert.Equal(t, int64(0), expiredAt)
	assert.True(t, found)
	assert.Nil(t, err)
}

func TestSet(t *testing.T) {
	provider, _ := New("sync-map")

	// set int value
	err := provider.Set("test", float64(100), 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set with int: %v", err)
	}
	value, _, _, _ := provider.Get("test")
	assert.Equal(t, float64(100), value)

	// set string value
	err = provider.Set("test", "100", 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set with string value: %v", err)
	}
	value, _, _, _ = provider.Get("test")
	assert.Equal(t, "100", value)

	// set array value
	valueArray := []interface{}{"1", "2"}
	err = provider.Set("test", valueArray, 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set with string value: %v", err)
	}
	value, _, _, _ = provider.Get("test")
	assert.Equal(t, valueArray, value)

	// set Map value
	valueMap := map[string]interface{}{"1": "5"}
	err = provider.Set("test", valueMap, 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set with string value: %v", err)
	}
	value, _, _, _ = provider.Get("test")
	assert.Equal(t, valueMap, value)

	// validate value expiration after 1 sec
	err = provider.Set("test", "value", 1)
	if err != nil {
		t.Fatalf("Error occurred while calling Set: %v", err)
	}

	value, expiredAt, found, err := provider.Get("test")
	assert.Equal(t, "value", value)
	assert.NotEqual(t, int64(0), expiredAt)
	assert.True(t, found)
	assert.Nil(t, err)

	// wait for timeout
	//NOTE: bad approach to wait, better to manipulate with time ... quick workaround
	time.Sleep(time.Second)
	// check again
	value, expiredAt, found, err = provider.Get("test")
	assert.Nil(t, value)
	assert.Equal(t, int64(0), expiredAt)
	assert.False(t, found)
	assert.Nil(t, err)

	// call set twice
	var valueString = "value"
	provider.Set("test", valueString, 0)
	v, _, _, _ := provider.Get("test")
	assert.Equal(t, valueString, v)
	// second call rewrite value and ttl
	valueString = "value_2"
	expectedTimestamp := time.Now().Add(time.Second * time.Duration(3600)).Unix()
	provider.Set("test", valueString, 3600)
	v, expiredAt, _, _ = provider.Get("test")
	assert.Equal(t, valueString, v)
	assert.LessOrEqual(t, expectedTimestamp, expiredAt)

}

func TestDelete(t *testing.T) {
	provider, _ := New("sync-map")

	// delete nonexistent keys
	_, _, found, _ := provider.Get("test")
	assert.False(t, found)
	err := provider.Delete("test")
	if err != nil {
		t.Fatalf("Error occurred while deleting nonexistent key: %v", err)
	}

	// delete existing key
	err = provider.Set("test", "value", 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set: %v", err)
	}

	err = provider.Delete("test")
	if err != nil {
		t.Fatalf("Error occurred while deleting record: %v", err)
	}
	value, expiredAt, found, err := provider.Get("test")
	assert.Nil(t, value)
	assert.Equal(t, int64(0), expiredAt)
	assert.False(t, found)
	assert.Nil(t, err)
}

func TestGetKeys(t *testing.T) {
	provider, _ := New("sync-map")
	// check that cache is empty
	keys, err := provider.GetKeys()
	if err != nil {
		t.Fatalf("Error occurred while getting keys record: %v", err)
	}
	assert.Equal(t, []string{}, keys)

	err = provider.Set("test", "value", 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set: %v", err)
	}

	err = provider.Set("test_array", "[1, 2, 3]", 0)
	if err != nil {
		t.Fatalf("Error occurred while calling Set: %v", err)
	}

	keys, err = provider.GetKeys()
	if err != nil {
		t.Fatalf("Error occurred while getting keys record: %v", err)
	}

	assert.Equal(t, []string{"test", "test_array"}, keys)
}
func BenchmarkGetMutexMap(b *testing.B) {
	provider, _ := New("mutex-map")
	provider.Set("test_int", 1, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Get("test_int")
	}
}

func BenchmarkSetMutexMap(b *testing.B) {
	provider, _ := New("mutex-map")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Set("test_int_"+string(i), i, 0)
	}
}

func BenchmarkDeleteMutexMap(b *testing.B) {
	provider, _ := New("mutex-map")
	for i := 0; i < b.N; i++ {
		provider.Set("test_int_"+string(i), i, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Delete("test_int_" + string(i))
	}
}

func BenchmarkGetKeysMutexMap(b *testing.B) {
	provider, _ := New("mutex-map")
	for i := 0; i < b.N; i++ {
		provider.Set("test_int_"+string(i), i, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.GetKeys()
	}
}

func BenchmarkGetSyncMap(b *testing.B) {
	provider, _ := New("sync-map")
	provider.Set("test_int", 1, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Get("test_int")
	}
}

func BenchmarkSetSyncMap(b *testing.B) {
	provider, _ := New("sync-map")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Set("test_int_"+string(i), i, 0)
	}
}

func BenchmarkDeleteSyncMap(b *testing.B) {
	provider, _ := New("sync-map")
	for i := 0; i < b.N; i++ {
		provider.Set("test_int_"+string(i), i, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Delete("test_int_" + string(i))
	}
}

func BenchmarkGetKeysSyncMap(b *testing.B) {
	provider, _ := New("sync-map")
	for i := 0; i < b.N; i++ {
		provider.Set("test_int_"+string(i), i, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.GetKeys()
	}
}
