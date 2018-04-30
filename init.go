package main

import (
	"flag"
)

func init() {
	flag.StringVar(&FlagFilters, "filters", "", "Filters to apply separated by ,")
	flag.StringVar(&FlagTemplate, "template", "{{ . | json }}", "Template to apply using text/template")
}
