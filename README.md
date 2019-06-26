# Cacher - A Go (golang) In-memory cache with Telnet/HTTP interface

## Build app
```
> ./Makefile cacher
> ./cacher --help
usage: cacher [<flags>]

In-memory Redis-like cache.

Flags:
      --help                    Show context-sensitive help (also try --help-long and --help-man).
  -i, --interface="http"        Either http or telnet interface enable.
  -a, --server=127.0.0.1        Server address.
  -p, --port="1323"             Server port.
      --auth_token=AUTH_TOKEN   Bearer Authentication Token.
  -t, --cache_type="mutex-map"  Select cache implementation.
      --version                 Show application version.
```

## Run HTTP server
```
> ./cacher -t sync-map --auth_token 0123456789
```

## Run Telnet server
```
> ./cacher -t sync-map -i telnet -p 5555
```

## Run test
```
> ./Makefile test
```

## Run benchmarks
```
> ./Makefile bench
```

## Build CLI
```
> ./Makefile cacher_cli
> ./cacher_cli --help  
usage: cacher-cli [<flags>] <command> [<args> ...]

CLI for Cacher application.

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Commands:
  help [<command>...]
    Show help.

  telnet [<host>] [<port>]
    Run Telnet client and connect with Telnet Cacher server

  http --auth_token=AUTH_TOKEN [<flags>] <command> [<key>] [<value>] [<ttl>]
    Use http client to send commands to Cacher.

> ./cacher -t sync-map --auth_token 0123456789
```

## Run HTTP server
```
> ./cacher_cli http -t 00000 set key {\"test\":10} 3600
> ./cacher_cli http -t 00000 get key
```

## Run Telnet client
```
./cacher_cli telnet
```


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

#### Get all cache keys
```
curl http://localhost:1323/keys -H 'Authorization: Bearer 0123456789'
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

#### Get all cache keys
```
> keys
```