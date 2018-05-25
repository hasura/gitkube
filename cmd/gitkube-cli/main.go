// A CLI tool to install and manage Gitkube and associated remotes on a
// Kubernetes cluster.
package main

import (
	"github.com/hasura/gitkube/pkg/cmd"
	log "github.com/sirupsen/logrus"
)

// main is the entrypoint function
func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
