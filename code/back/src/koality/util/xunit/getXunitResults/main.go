package main

import (
	"koality/util/xunit"
	"os"
)

func main() {
	xunit.GetXunitResults(os.Args[1], os.Args[2:], os.Stdout, os.Stderr)
}
