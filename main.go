package main

import (
	"github.com/docker/docker/api/types"
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/filters"
	"github.com/tehmoon/errors"
	"encoding/json"
	"fmt"
	"os"
	"flag"
	"strings"
	"net/url"
)

var FlagFilters string

func init() {
	flag.StringVar(&FlagFilters, "filters", "", "Filters to apply separated by ,")
}

func parseFilters(str string) (*filters.Args, error) {
	args := filters.NewArgs()

	fs := strings.Split(str, ",")
	for _, f := range fs {
		values, err := url.ParseQuery(f)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing filter: %s", f)
		}

		for k, v := range values {
			args.Add(k, v[0])
		}
	}

	return &args, nil
}

func main() {
	flag.Parse()

	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, errors.Wrap(err, "Error creating new docker client").Error())
		os.Exit(1)
	}

	myFilters, err := parseFilters(FlagFilters)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Error parsing filters").Error())
		os.Exit(1)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: *myFilters,
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
