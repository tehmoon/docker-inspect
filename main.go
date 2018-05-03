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
	"net/url"
	"os/exec"
)

var FlagFilters = ValueFlagStringArray{}

type ValueFlagStringArray []string

func (vfsa ValueFlagStringArray) String() (string) {
	return ""
}

func (vfsa *ValueFlagStringArray) Set(value string) (error) {
	*vfsa = append(*vfsa, value)

	return nil
}

func init() {
	flag.Var(&FlagFilters, "filter", "Filter to pass to docker. Can be repeated.")
}

func newFilters(fs []string) (*filters.Args, error) {
	args := filters.NewArgs()

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

func getApiVersion() (string, error) {
	cmd := exec.Command("docker", "version", "-f {{ .Server.MinAPIVersion | json }}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "Combined Output: \t%s", string(output[:]))
	}

	return string(output[2:len(output) - 2]), nil
}

func setVersionEnv() (error) {
	if version := os.Getenv("DOCKER_API_VERSION"); version != "" {
		return nil
	}

	version, err := getApiVersion()
	if err != nil {
		return errors.Wrap(err, "Error getting the version of docker server")
	}

	err = os.Setenv("DOCKER_API_VERSION", version)
	if err != nil {
		return errors.Wrap(err, "Error setting the environment DOCKER_API_VERSION")
	}

	return nil
}

func main() {
	flag.Parse()

	err := setVersionEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Error creating new docker client").Error())
		os.Exit(1)
	}

	myFilters, err := newFilters(FlagFilters)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Error building filters").Error())
		os.Exit(1)
	}

	containers, err := inspectContainers(cli, myFilters)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	payload, err := json.Marshal(containers)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Error marshaling to json").Error())
		os.Exit(1)
	}

	fmt.Println(string(payload[:]))
}

func inspectContainers(cli *client.Client, myFilters *filters.Args) ([]types.ContainerJSON, error) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: *myFilters,
	})

	if err != nil {
		return nil, errors.Wrap(err, "Error listing containers")
	}

	cjs := make([]types.ContainerJSON, 0)

	for _, container := range containers {
		cj, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "Error inspecting container %d", container.ID)
		}

		cjs = append(cjs, cj)
	}

	return cjs, nil
}
