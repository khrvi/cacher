# Cacher - A Go (golang) In-memory cache with Telnet/HTTP interface

## HTTP interface (CURL examples):

#### Set new value of string:
```
curl \
  -X POST \
  http://localhost:1323/ \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer 0123456789' \
  -d '{"key":"test_string","value":"test_value", "ttl": 30}'
```

#### Set new value of array:
```
curl \
  -X POST \
  http://localhost:1323/ \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer 0123456789' \
  -d '{"key":"test_array","value":["one","two","three"]}'
```

#### Set new value of map:
```
curl \
  -X POST \
  http://localhost:1323/ \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer 0123456789' \
  -d '{"key":"test_map","value":{"test":1}}'
```

## Telnet interface:
```
> telnet localhost 5555
```

#### Set operation: set <key> <value> [<ttl>]
Note: whitespace is used as params separator
#### Set new value of string: 
```
> set test_string test_value
```

#### Set new value of array: 
```
> set test_array [1,2,3] 50
```

#### Set new value of map: 
```
> set test_map {"test":1} 3600
```