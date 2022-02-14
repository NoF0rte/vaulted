package main

import (
	"errors"
	"fmt"

	vaulted "github.com/NoF0rte/vaulted/v3/lib"
)

type Remove struct {
	VaultNames []string
}

func (r *Remove) Run(store vaulted.Store) error {
	failures := 0
	for _, name := range r.VaultNames {
		err := store.RemoveVault(name)
		if err != nil {
			failures++
			fmt.Printf("%s: %v\n", name, err)
		}
	}

	if failures > 0 {
		return ErrorWithExitCode{
			errors.New("Vault could not be removed"),
			failures,
		}
	}

	return nil
}
