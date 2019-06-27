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
      --cdb                     Enable or disable save on disk using CDB.
      --cdb_period=60           Period in seconds of dumping data to CDB.
      --appendonly              Enable or disable Append-only file.
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
```

## Run HTTP server
```
> ./cacher -t sync-map --auth_token 0123456789
```

## Cacher Persistence
Cacher persistance implemented using Redis similar approach. There two options how persistance can be provided.

#### CDB
CDB is implemented by snapshoting in-memory cache to disk at specified intervals. By default CDB is enabled with interval 60 secs.
Managing CDB and dumping interval could be changed using cli flags.
In case if CDB is enabled and interval is 0 then every single Write/Delete Cacher operation will be immediately saved to CDB.
In case if CDB is enabled and interval is greater than 0 then Batch Write to disk will be performed.
CDB is using LevelDB as key-value disk storage.

In order to disable dumping data to disk could be used the following command:
```
> ./cacher -i telnet -p 5555 --no-cdb
```

#### Append-only file
AOF works using transaction log where every SET and DELETE operations will be appended to file. It should help to restore
operations that were initiated but never finished. For example after killing process with 'kill -9'. Situations like CTRL+C
handled separatelly by catching this signal and hence CDB batch operation has to finish dumping before exit.
Command to disable AOF:
```
> ./cacher -i telnet -p 5555 --no-appendonly
```

#### CDB + AOF
Both modes could work together to provide higher durability. If both modes are enabled then restore from CDB processed and after that restore from AOF.
AOF will restore only commands missed by CDB. In case if only AOF enabled then a full AOF log will be restored to in-memory DB.


## CLI examples
#### Run NTTP client
Request for value of 'test' key 
```
> ./cacher_cli http get test --auth_token 00000
```
Set array value
```
> ./cacher_cli http set test_array [9,2] --auth_token 00000
```
Set string value
```
> ./cacher_cli http set test_string \"string\" --auth_token 00000
```

#### Run Telnet client
Telnet client works as Standard telnet client in interactive mode.
Run client
```
> ./cacher_cli telnet
> set test [1,2] 30
> keys
> get test
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