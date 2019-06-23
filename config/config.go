package config

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "1.0.0"
	app     = kingpin.New("cacher", "In-memory Redis-like cache.")

	Interface = app.Flag("interface", "Either http or telnet interface enable.").
			Short('i').
			Default("http").
			HintOptions("http", "telnet").
			String()

	ServerIP   = app.Flag("server", "Server address.").Short('a').Default("127.0.0.1").IP()
	ServerPort = app.Flag("port", "Server port.").Short('p').Default("1323").String()
	AuthToken  = app.Flag("auth_token", "Bearer Authentication Token.").String()

	// cache options
	CacheType = app.Flag("cache_type", "Select cache implementation.").
			Short('t').
			Default("mutex-map").
			HintOptions("mutex-map", "sync-map").
			String()
)

func init() {
	// kingpin.Parse()
	// Parse command line
	app.Version(version)
	app.Parse(os.Args[1:])
	switch *Interface {
	case "http":
		if *AuthToken == "" {
			kingpin.Fatalf("Please set 'auth_token' option first to start server! %s", *AuthToken)
		}

	case "telnet":
		// telnet
	default:
		kingpin.Fatalf("Unknown Interface type: %s", *Interface)
	}
}
