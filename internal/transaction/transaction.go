package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

const subsidy = 10

// Transaction represents a blockchain transaction
type Transaction struct {
	ID   []byte
	Data string
	Vin  []TXInput
	Vout []TXOutput
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// TXOutput represents a transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

// CanUnlockOutputWith checks whether the address initiated the transaction
func (in *TXInput) CanUnlockOutputWith(address string) bool {
	// Get address from public key
	pubKeyHash := HashPubKey(in.PubKey)
	addressFromPubKey := getAddressFromPubKey(pubKeyHash)
	return addressFromPubKey == address
}

// CanBeUnlockedWith checks if the output can be unlocked with the provided data
func (out *TXOutput) CanBeUnlockedWith(address string) bool {
	return out.ScriptPubKey == address
}

// TXOutputMap is a map of transaction ID to output indices
type TXOutputMap map[string][]int

// UTXOSearcher interface for finding spendable outputs
type UTXOSearcher interface {
	FindSpendableOutputs(address string, amount int) (int, TXOutputMap)
	FindUTXO(address string) []TXOutput
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, nil, nil}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, data, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, bc UTXOSearcher, privKey *ecdsa.PrivateKey, pubKey []byte) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		return nil, fmt.Errorf("not enough funds: address '%s' has balance %d, but tried to send %d", from, acc, amount)
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, fmt.Errorf("failed to decode transaction id: %v", err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from}) // a change
	}

	tx := Transaction{nil, "", inputs, outputs}
	tx.SetID()

	// Sign the transaction with the private key
	if privKey != nil {
		tx.Sign(privKey)
	}

	return &tx, nil
}

// Sign signs the transaction with the given private key
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey) {
	// First, create a trimmed copy of the transaction for signing
	txCopy := tx.TrimmedCopy()

	// Sign each input
	for i := range txCopy.Vin {
		txCopy.Vin[i].Signature = nil
		txCopy.Vin[i].PubKey = nil

		dataToSign := txCopy.getDataToSign(i)
		r, s, err := ecdsa.Sign(rand.Reader, privKey, dataToSign)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[i].Signature = signature
	}
}

// TrimmedCopy creates a trimmed copy of the transaction for signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{
			Txid:      vin.Txid,
			Vout:      vin.Vout,
			Signature: nil,
			PubKey:    nil,
		})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{
			Value:        vout.Value,
			ScriptPubKey: vout.ScriptPubKey,
		})
	}

	return Transaction{
		ID:   tx.ID,
		Data: tx.Data,
		Vin:  inputs,
		Vout: outputs,
	}
}

// getDataToSign returns the data to sign for a given input
func (tx *Transaction) getDataToSign(inputIndex int) []byte {
	txCopy := tx.TrimmedCopy()
	txCopy.Vin[inputIndex].Signature = nil
	txCopy.Vin[inputIndex].PubKey = nil

	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(txCopy)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(data.Bytes())
	return hash[:]
}

// Verify verifies all transaction input signatures
func (tx *Transaction) Verify() bool {
	txCopy := tx.TrimmedCopy()

	for i, vin := range tx.Vin {
		dataToVerify := txCopy.getDataToSign(i)

		pubKey := vin.PubKey
		signature := vin.Signature

		// Parse signature
		r := new(big.Int)
		s := new(big.Int)
		sigLen := len(signature) / 2
		r.SetBytes(signature[:sigLen])
		s.SetBytes(signature[sigLen:])

		x := new(big.Int).SetBytes(pubKey[:len(pubKey)/2])
		y := new(big.Int).SetBytes(pubKey[len(pubKey)/2:])

		ecdsaPubKey := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		}

		if !ecdsa.Verify(ecdsaPubKey, dataToVerify, r, s) {
			return false
		}
	}

	return true
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	hash := sha256.Sum256(pubKey)
	return hash[:]
}

// getAddressFromPubKey generates address from public key hash
func getAddressFromPubKey(pubKeyHash []byte) string {
	versionedPayload := append([]byte{byte(0x00)}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	return base58Encode(fullPayload)
}

// checksum generates checksum for address
func checksum(payload []byte) []byte {
	hash := sha256.Sum256(payload)
	return hash[:4]
}

// base58Encode encodes a byte slice to base58
func base58Encode(input []byte) string {
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
		return string(alphabet[0])
	}

	return string(result)
}
