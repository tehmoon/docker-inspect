package main

import (
	"github.com/docker/docker/api/types"
	"context"
	"github.com/docker/docker/client"
	"github.com/tehmoon/errors"
	"encoding/json"
	"fmt"
	"os"
	"flag"
)

func main() {
	flag.Parse()

	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, errors.Wrap(err, "Error creating new docker client").Error())
		os.Exit(1)
	}

	dockerFilters, err := parseFilters(FlagFilters)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Error parsing filters"))
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: *dockerFilters,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	payload, err := json.Marshal(containers)
	if err != nil {
		fmt.Fprintf(os.Stderr, errors.Wrap(err, "Error marshaling to json").Error())
		os.Exit(1)
	}

	fmt.Println(string(payload[:]))
}
