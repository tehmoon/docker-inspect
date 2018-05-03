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
	"text/template"
	"bytes"
	"io"
)

var (
	FlagFilters = ValueFlagStringArray{}
	FlagTemplates = ValueFlagStringArray{}
	jsonFunc = template.FuncMap{
		"json": func(d interface{}) string {
			payload, err := json.Marshal(d)
			if err != nil {
				return ""
			}

			return string(payload[:])
		},
	}
)

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
	flag.Var(&FlagTemplates, "template", "JSON template to pass to docker inspect")
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

	templates, err := newTemplates(FlagTemplates)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = outputTemplates(templates, containers, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
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

func newTemplates(templates []string) ([]*template.Template, error) {
	tt := make([]*template.Template, 0)

	if len(templates) == 0 {
		templates = append(templates, "{{ . | json }}")
	}

	for i, tmpl := range templates {
		name := fmt.Sprintf("main_%d", i)

		t, err := template.New(name).Funcs(jsonFunc).Parse(tmpl)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing template")
		}

		tt = append(tt, t)
	}

	return tt, nil
}

func outputTemplate(tmpl *template.Template, containers []types.ContainerJSON, writer io.Writer) (error) {
	enc := json.NewEncoder(writer)

	array := make([]interface{}, 0)
	buf := bytes.NewBuffer([]byte{})

	for _, container := range containers {
		var v interface{}
		buf.Reset()

		err := tmpl.Execute(buf, container)
		if err != nil {
			return errors.Wrap(err, "Error executing the template")
		}

		err = json.Unmarshal(buf.Bytes(), &v)
		if err != nil {
			return errors.Wrap(err, "Error marshaling output template to JSON")
		}

		array = append(array, v)
	}

	err := enc.Encode(array)
	if err != nil {
		return errors.Wrap(err, "Error marshaling to json")
	}

	return nil
}

func outputTemplates(templates []*template.Template, containers []types.ContainerJSON, writer io.Writer) (error) {
	var err error

	for _, tmpl := range templates {
		err = outputTemplate(tmpl, containers, writer)
		if err != nil {
			return err
		}
	}

	return nil
}
