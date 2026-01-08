package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/aspertheghost/blockchain/internal/blockchain"
	"github.com/aspertheghost/blockchain/internal/transaction"
)

const walletFile = "wallet.dat"

// Wallet holds private and public keys
type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

// Wallets holds a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallet creates and returns a new Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// GetAddress generates a wallet address from the public key
func (w *Wallet) GetAddress() string {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{byte(0x00)}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := base58Encode(fullPayload)

	return string(address)
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)
	return pubHash[:]
}

// checksum generates checksum for address
func checksum(payload []byte) []byte {
	hash := sha256.Sum256(payload)
	return hash[:4]
}

// newKeyPair generates a new key pair
func newKeyPair() ([]byte, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pubKey := append(private.X.Bytes(), private.Y.Bytes()...)

	return private.D.Bytes(), pubKey
}

// base58Encode encodes a byte slice to base58
func base58Encode(input []byte) []byte {
	alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	var result []byte
	x := new(big.Int)
	x.SetBytes(input)

	base := big.NewInt(int64(len(alphabet)))
	zero := big.NewInt(0)

	for x.Cmp(zero) != 0 {
		mod := new(big.Int)
		x.DivMod(x, base, mod)
		result = append(result, alphabet[mod.Int64()])
	}

	// Reverse result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	if len(result) == 0 {
		return []byte{alphabet[0]}
	}

	return result
}

// NewWallets creates a Wallets object and loads wallets from file
func NewWallets() (*Wallets, error) {
	ws := &Wallets{}
	ws.Wallets = make(map[string]*Wallet)

	err := ws.LoadFromFile()

	return ws, err
}

// CreateWallet creates a new wallet and returns its address
func (ws *Wallets) CreateWallet() string {
	return ws.CreateWalletWithName("")
}

// CreateWalletWithName creates a new wallet with an optional name alias
func (ws *Wallets) CreateWalletWithName(name string) string {
	wallet := NewWallet()
	address := wallet.GetAddress()

	ws.Wallets[address] = wallet

	// Save the wallet to file
	ws.SaveToFile()

	// Fund the wallet with 10 coins
	err := FundWallet(address)
	if err != nil {
		fmt.Printf("Warning: Could not fund wallet with initial 10 coins: %v\n", err)
	}

	// If a name is provided, save the alias
	if name != "" {
		aliases, err := NewWalletAliases()
		if err == nil {
			aliases.SetAlias(name, address)
			aliases.SaveToFile()
			fmt.Printf("Wallet '%s' created with address: %s\n", name, address)
		} else {
			fmt.Printf("Wallet created with address: %s (warning: could not save alias: %v)\n", address, err)
		}
	} else {
		fmt.Printf("New wallet created with address: %s\n", address)
	}

	return address
}

// FundWallet adds 10 coins to a wallet by creating a coinbase transaction
func FundWallet(address string) error {
	// Check if blockchain exists
	if !blockchain.DBExists() {
		// Create new blockchain with this wallet as the genesis block recipient
		bc := blockchain.CreateBlockchain(address)
		bc.Close()
		fmt.Printf("Created new blockchain with genesis block reward of 10 coins for %s\n", address)
		return nil
	}

	// Blockchain exists, create a new block with a coinbase transaction
	bc := blockchain.NewBlockchain(address)
	defer bc.Close()

	// Create a coinbase transaction to the wallet
	cbTx := transaction.NewCoinbaseTX(address, "Initial reward for new wallet")

	// Mine the block
	bc.MineBlock([]*transaction.Transaction{cbTx})
	fmt.Printf("Added 10 coins to wallet %s\n", address)

	return nil
}

// GetWallet returns wallet for a given address
func (ws *Wallets) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}

// SaveToFile saves wallets to a file
func (ws *Wallets) SaveToFile() {
	var content bytes.Buffer

	enc := gob.NewEncoder(&content)
	err := enc.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// LoadFromFile loads wallets from a file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return nil
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	dec := gob.NewDecoder(bytes.NewReader(fileContent))
	err = dec.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}
