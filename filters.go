package main

import (
	"net/url"
	"strings"
	"github.com/tehmoon/errors"
	"github.com/docker/docker/api/types/filters"
)

var FlagFilters string
var FlagTemplate string

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
