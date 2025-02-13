package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	vaulted "github.com/NoF0rte/vaulted/v3/lib"
)

type Load struct {
	VaultName string
}

func (l Load) Run(store vaulted.Store) error {
	jvault, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	vault := &vaulted.Vault{}
	err = json.Unmarshal(jvault, vault)
	if err != nil {
		return err
	}

	err = store.SealVault(vault, l.VaultName)
	if err != nil {
		return err
	}

	return nil
}
