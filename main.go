package main

import (
	// "github.com/hugopicq/cartographergo/cartographer"
	// "github.com/hugopicq/cartographergo/cartographer/modules"
	"github.com/hugopicq/cartographergo/cmd"
)

func main() {
	cmd.Execute()
	// test()
}

// func test() {
// 	creds := cartographer.Credentials{
// 		Domain:           "hackcorp.local",
// 		DomainController: "192.168.230.10",
// 		User:             "alice",
// 		Password:         "Client1!",
// 	}

// 	module := new(modules.SessionsModule)

// 	module.Prepare(&creds)

// }
