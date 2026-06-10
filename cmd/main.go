package main

import (
	"fmt"

	"github.com/scality/nodeip-discovery/cmd/config"
)

func main() {
	fmt.Printf("Starting %s:%s\n", config.ApplicationName, config.ApplicationVersion)
}
