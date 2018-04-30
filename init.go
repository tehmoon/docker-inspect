package main

import (
	"flag"
)

func init() {
  flag.StringVar(&FlagFilters, "filters", "", "Filters to apply separated by ,")
}
