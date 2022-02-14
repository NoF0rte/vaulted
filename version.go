package main

import (
	"fmt"

	vaulted "github.com/NoF0rte/vaulted/v3/lib"
)

const (
	VERSION = "3.1.unstable"
)

type Version struct{}

func (l *Version) Run(store vaulted.Store) error {
	fmt.Printf("Vaulted v%s\n", VERSION)
	return nil
}
