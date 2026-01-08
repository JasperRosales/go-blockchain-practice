package wallet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

const aliasesFile = "wallet-aliases.dat"

// WalletAliases holds name-to-address mappings
type WalletAliases struct {
	Aliases map[string]string // name -> address
}

// NewWalletAliases creates a new WalletAliases object and loads from file
func NewWalletAliases() (*WalletAliases, error) {
	wa := &WalletAliases{}
	wa.Aliases = make(map[string]string)

	err := wa.LoadFromFile()

	return wa, err
}

// SetAlias creates a name-to-address mapping
func (wa *WalletAliases) SetAlias(name, address string) {
	wa.Aliases[name] = address
}

// GetAddress returns the address for a given name
func (wa *WalletAliases) GetAddress(name string) string {
	if addr, ok := wa.Aliases[name]; ok {
		return addr
	}
	return name // Return original name if not an alias
}

// GetAlias returns the name for a given address
func (wa *WalletAliases) GetAlias(address string) string {
	for name, addr := range wa.Aliases {
		if addr == address {
			return name
		}
	}
	return "" // Return empty if no alias found
}

// SaveToFile saves aliases to a file
func (wa *WalletAliases) SaveToFile() {
	var content bytes.Buffer

	enc := gob.NewEncoder(&content)
	err := enc.Encode(wa)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(aliasesFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// LoadFromFile loads aliases from a file
func (wa *WalletAliases) LoadFromFile() error {
	if _, err := os.Stat(aliasesFile); os.IsNotExist(err) {
		return nil
	}

	fileContent, err := os.ReadFile(aliasesFile)
	if err != nil {
		log.Panic(err)
	}

	var aliases WalletAliases
	dec := gob.NewDecoder(bytes.NewReader(fileContent))
	err = dec.Decode(&aliases)
	if err != nil {
		return err
	}

	wa.Aliases = aliases.Aliases

	return nil
}

// PrintAliases prints all name-to-address mappings
func (wa *WalletAliases) PrintAliases() {
	fmt.Println("Wallet Aliases:")
	for name, addr := range wa.Aliases {
		fmt.Printf("  %s -> %s\n", name, addr)
	}
}
