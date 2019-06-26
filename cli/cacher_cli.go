package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ddliu/go-httpclient"
	"github.com/reiver/go-telnet"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "1.0.0"
	app     = kingpin.New("cacher-cli", "CLI for Cacher application.")

	telnetMode = app.Command("telnet", "Run Telnet client and connect with Telnet Cacher server")
	host       = telnetMode.Arg("host", "Server host.").Default("127.0.0.1").IP()
	port       = telnetMode.Arg("port", "Server port.").Default("5555").String()

	http       = app.Command("http", "Use http client to send commands to Cacher.")
	serverIP   = http.Flag("server", "Server address.").Short('a').Default("127.0.0.1").IP()
	serverPort = http.Flag("port", "Server port.").Short('p').Default("1323").String()
	authToken  = http.Flag("auth_token", "Bearer Authentication Token.").Short('t').Required().String()
	command    = http.Arg("command", "Command for Cacher.").HintOptions("get", "set", "delete", "keys").Required().String()
	key        = http.Arg("key", "Cache key.").String()
	value      = http.Arg("value", "Cache value. Should be in JSON form. Only for 'set' command!").String()
	ttl        = http.Arg("ttl", "Cache pait TTL in seconds.").Int64()
)

type payload struct {
	Key   string      `json:"key" form:"key" query:"key"`
	Value interface{} `json:"value" form:"value" query:"value"`
	TTL   int64       `json:"ttl" form:"ttl" query:"ttl"`
}

func main() {
	app.Version(version)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Telnet client mode
	case telnetMode.FullCommand():
		startTelNetClient()

	// Http mode
	case http.FullCommand():
		httpclient.Defaults(httpclient.Map{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + *authToken,
		})
		if *command == "get" {
			if *key == "" {
				kingpin.Fatalf("Command 'get' requires one param: 'key'.")
			}
			handleGetCommand()
		} else if *command == "set" {
			if *key == "" || *value == "" {
				kingpin.Fatalf("Command 'set' requires at least two params: 'key' and 'value'. Third param 'ttl' is optinal.")
			}
			handleSetCommand()
		} else if *command == "delete" {
			if *key == "" {
				kingpin.Fatalf("Command 'delete' requires one param: 'key'.")
			}
			handleDeleteCommand()
		} else if *command == "keys" {
			handleKeysCommand()
		}
	}
}

func startTelNetClient() {
	address := fmt.Sprintf("%s:%s", *host, *port)
	fmt.Printf("Connecting to %s...\n", address)
	var caller = telnet.StandardCaller
	err := telnet.DialToAndCall(address, caller)
	if err != nil {
		panic(err)
	}
}

func handleGetCommand() {
	url := fmt.Sprintf("http://%s:%s/%s", *serverIP, *serverPort, *key)
	handleHTTPResponse(httpclient.Get(url))
}

func handleSetCommand() {
	url := fmt.Sprintf("http://%s:%s/", *serverIP, *serverPort)

	var rawValue interface{}
	err := json.Unmarshal([]byte(*value), &rawValue)
	if err != nil {
		kingpin.Fatalf("Error occurred while unmarshing value '%s': %+v", *value, err)
	}
	data := payload{
		Key:   *key,
		Value: rawValue,
		TTL:   *ttl,
	}
	body, err := json.Marshal(data)
	if err != nil {
		kingpin.Fatalf("Error occurred while marshing payload %+v: %+v", data, err)
	}

	handleHTTPResponse(httpclient.PostJson(url, string(body)))
}

func handleDeleteCommand() {
	url := fmt.Sprintf("http://%s:%s/%s", *serverIP, *serverPort, *key)
	handleHTTPResponse(httpclient.Delete(url))
}

func handleKeysCommand() {
	url := fmt.Sprintf("http://%s:%s/keys", *serverIP, *serverPort)
	handleHTTPResponse(httpclient.Get(url))
}

func handleHTTPResponse(res *httpclient.Response, err error) {
	if err != nil {
		kingpin.Fatalf("Error occurred while getting key '%s': %+v", *key, err)
	}
	bodyString, err := res.ToString()
	if err != nil {
		kingpin.Fatalf("Error occurred while getting key '%s': %+v", *key, err)
	}
	println(bodyString)
}
