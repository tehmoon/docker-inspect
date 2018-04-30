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
	"io"
	"text/template"
)

var jsonFunc = template.FuncMap{
	"json": func(d interface{}) string {
		payload, err := json.Marshal(d)
		if err != nil {
			return ""
		}

		return string(payload[:])
	},
}

func main() {
	flag.Parse()

	tmpl, err := template.New("main").Funcs(jsonFunc).Parse(FlagTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

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

	reader, writer := io.Pipe()
	dec := json.NewDecoder(reader)

	go func() {
		var err error

		for _, container := range containers {
			err = tmpl.Execute(writer, container)
			if err != nil {
				break
			}
		}

		writer.CloseWithError(err)
	}()

	output := make([]interface{}, 0)

	for {
		var v interface{}

		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		output = append(output, v)
	}

	payload, err := json.Marshal(&output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(string(payload[:]))
}
