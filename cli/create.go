package cli

import (
	"fmt"
	"log"

	"github.com/docker/swarm/discovery/token"
	"github.com/urfave/cli"
)

func create(c *cli.Context) {
	
	if len(c.Args()) != 0 {
		log.Fatalf("the `create` command takes no arguments. See '%s create --help'.", c.App.Name)
	}
	discovery := &token.Discovery{}
	discovery.Initialize("", 0, 0, nil)
	token, err := discovery.CreateCluster()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("BUGA")
	fmt.Println(token)
}
