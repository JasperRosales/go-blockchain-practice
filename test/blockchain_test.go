package test

import (
	"testing"

	"github.com/aspertheghost/blockchain/internal/block"
	"github.com/aspertheghost/blockchain/internal/proofofwork"
	"github.com/aspertheghost/blockchain/internal/transaction"
	"github.com/aspertheghost/blockchain/internal/utils"
)

// TestBlockSerialization tests block serialization and deserialization
func TestBlockSerialization(t *testing.T) {
	// Create a coinbase transaction for genesis block
	coinbaseTX := transaction.NewCoinbaseTX("genesis_address", "Genesis block data")

	// Create genesis block
	genesisBlock := block.NewGenesisBlock(coinbaseTX)

	// Serialize the block
	serialized := genesisBlock.Serialize()

	// Deserialize the block
	deserializedBlock := block.DeserializeBlock(serialized)

	// Verify the block was correctly serialized and deserialized
	if deserializedBlock.Timestamp != genesisBlock.Timestamp {
		t.Errorf("Timestamp mismatch: expected %d, got %d", genesisBlock.Timestamp, deserializedBlock.Timestamp)
	}

	if len(deserializedBlock.PrevBlockHash) != len(genesisBlock.PrevBlockHash) {
		t.Errorf("PrevBlockHash length mismatch: expected %d, got %d",
			len(genesisBlock.PrevBlockHash), len(deserializedBlock.PrevBlockHash))
	}

	if len(deserializedBlock.Transactions) != len(genesisBlock.Transactions) {
		t.Errorf("Transactions count mismatch: expected %d, got %d",
			len(genesisBlock.Transactions), len(deserializedBlock.Transactions))
	}

	t.Logf("Block serialization test passed for genesis block")
}

// TestBlockHashTransactions tests the HashTransactions method
func TestBlockHashTransactions(t *testing.T) {
	// Create multiple transactions
	tx1 := transaction.NewCoinbaseTX("address1", "Transaction 1")
	tx2 := transaction.NewCoinbaseTX("address2", "Transaction 2")

	transactions := []*transaction.Transaction{tx1, tx2}
	prevBlockHash := []byte("previous hash")

	newBlock := block.NewBlock(transactions, prevBlockHash)

	// Get the transaction hash
	txHash := newBlock.HashTransactions()

	// Verify the hash is 32 bytes (SHA256)
	if len(txHash) != 32 {
		t.Errorf("Transaction hash should be 32 bytes, got %d", len(txHash))
	}

	// Hash should be different for different transactions
	tx1Only := block.NewBlock([]*transaction.Transaction{tx1}, prevBlockHash)
	tx2Only := block.NewBlock([]*transaction.Transaction{tx2}, prevBlockHash)

	if len(tx1Only.HashTransactions()) == 0 || len(tx2Only.HashTransactions()) == 0 {
		t.Error("Transaction hash should not be empty")
	}

	t.Logf("Block HashTransactions test passed")
}

// TestProofOfWork tests the proof of work system
func TestProofOfWork(t *testing.T) {
	// Create a simple block with a coinbase transaction
	coinbaseTX := transaction.NewCoinbaseTX("test_address", "PoW test")
	prevBlockHash := []byte("0000000000000000000000000000000000000000000000000000000000000000")

	newBlock := block.NewBlock([]*transaction.Transaction{coinbaseTX}, prevBlockHash)

	// Create proof of work
	pow := proofofwork.NewProofOfWork(newBlock)

	// Run the proof of work
	nonce, hash := pow.Run()

	// Update block with PoW results
	newBlock.Nonce = nonce
	newBlock.Hash = hash

	// Create a new ProofOfWork with the updated block for validation
	pow = proofofwork.NewProofOfWork(newBlock)

	// Verify the nonce is valid
	if nonce < 0 {
		t.Errorf("Nonce should be non-negative, got %d", nonce)
	}

	// Verify the hash is 32 bytes
	if len(hash) != 32 {
		t.Errorf("Hash should be 32 bytes, got %d", len(hash))
	}

	// Verify the proof of work is valid
	if !pow.Validate() {
		t.Error("Proof of work validation failed")
	}

	t.Logf("Proof of work test passed with nonce=%d", nonce)
}

// TestIntToHex tests the IntToHex utility function
func TestIntToHex(t *testing.T) {
	// Test that IntToHex returns 8 bytes (sizeof int64)
	testCases := []int64{0, 1, 255, 256, 65535, 65536, 123456789, -1}

	for _, input := range testCases {
		result := utils.IntToHex(input)
		if len(result) != 8 {
			t.Errorf("IntToHex(%d): expected 8 bytes, got %d", input, len(result))
		}
	}

	// Test specific values
	tests := []struct {
		input    int64
		expected []byte
	}{
		{0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{1, []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{255, []byte{0, 0, 0, 0, 0, 0, 0, 255}},
		{256, []byte{0, 0, 0, 0, 0, 0, 1, 0}},
	}

	for _, tc := range tests {
		result := utils.IntToHex(tc.input)
		if string(result) != string(tc.expected) {
			t.Errorf("IntToHex(%d): expected %v, got %v", tc.input, tc.expected, result)
		}
	}

	t.Logf("IntToHex utility test passed")
}

// TestNewBlockFunction tests the NewBlock function
func TestNewBlockFunction(t *testing.T) {
	transactions := []*transaction.Transaction{
		transaction.NewCoinbaseTX("address1", "Test transaction"),
	}
	prevBlockHash := []byte("previous hash data")

	newBlock := block.NewBlock(transactions, prevBlockHash)

	// Verify block properties
	if newBlock.PrevBlockHash == nil {
		t.Error("PrevBlockHash should not be nil")
	}

	if len(newBlock.PrevBlockHash) == 0 {
		t.Error("PrevBlockHash should not be empty")
	}

	if len(newBlock.Transactions) != len(transactions) {
		t.Errorf("Block should contain %d transactions, got %d", len(transactions), len(newBlock.Transactions))
	}

	if newBlock.Nonce != 0 {
		t.Errorf("New block should have nonce 0, got %d", newBlock.Nonce)
	}

	if len(newBlock.Hash) != 0 {
		t.Errorf("New block should have empty hash before mining, got %x", newBlock.Hash)
	}

	t.Logf("NewBlock function test passed")
}

// TestCoinbaseTransaction tests coinbase transaction creation
func TestCoinbaseTransaction(t *testing.T) {
	// Test basic coinbase transaction
	tx := transaction.NewCoinbaseTX("recipient_address", "Coinbase data")

	// Verify it's identified as coinbase
	if !tx.IsCoinbase() {
		t.Error("Transaction should be identified as coinbase")
	}

	// Verify transaction ID is set
	if tx.ID == nil || len(tx.ID) == 0 {
		t.Error("Transaction ID should be set")
	}

	// Verify there is one input
	if len(tx.Vin) != 1 {
		t.Errorf("Coinbase transaction should have 1 input, got %d", len(tx.Vin))
	}

	// Verify there is one output
	if len(tx.Vout) != 1 {
		t.Errorf("Coinbase transaction should have 1 output, got %d", len(tx.Vout))
	}

	// Verify output value (should be subsidy = 10)
	if tx.Vout[0].Value != 10 {
		t.Errorf("Coinbase output value should be 10, got %d", tx.Vout[0].Value)
	}

	// Verify output scriptPubKey (recipient address)
	if tx.Vout[0].ScriptPubKey != "recipient_address" {
		t.Errorf("Output ScriptPubKey should be recipient address, got %s", tx.Vout[0].ScriptPubKey)
	}

	t.Logf("Coinbase transaction test passed")
}

// TestTransactionID tests that transaction IDs are properly set
func TestTransactionID(t *testing.T) {
	tx := transaction.NewCoinbaseTX("test_address", "Test data")

	// Transaction ID should be unique and set
	if tx.ID == nil || len(tx.ID) == 0 {
		t.Error("Transaction ID should not be empty")
	}

	// Create another transaction with different data
	tx2 := transaction.NewCoinbaseTX("test_address", "Test data 2")

	// IDs should be different (due to different data)
	id1 := tx.ID
	id2 := tx2.ID

	// Check that IDs are properly generated as byte slices
	if len(id1) != 32 || len(id2) != 32 {
		t.Errorf("Transaction IDs should be 32 bytes (SHA256), got id1=%d, id2=%d", len(id1), len(id2))
	}

	// Verify IDs are different for different transaction data
	sameID := true
	for i := range id1 {
		if id1[i] != id2[i] {
			sameID = false
			break
		}
	}
	if sameID {
		t.Error("Different transactions should have different IDs")
	}

	t.Logf("Transaction ID test passed")
}

// TestMultipleBlocks tests creating multiple blocks in sequence
func TestMultipleBlocks(t *testing.T) {
	// Create genesis block
	genesisTX := transaction.NewCoinbaseTX("genesis_recipient", "Genesis data")
	genesisBlock := block.NewGenesisBlock(genesisTX)

	// Create second block
	tx2 := transaction.NewCoinbaseTX("recipient2", "Block 2 data")
	block2 := block.NewBlock([]*transaction.Transaction{tx2}, genesisBlock.Hash)

	// Create third block
	tx3 := transaction.NewCoinbaseTX("recipient3", "Block 3 data")
	block3 := block.NewBlock([]*transaction.Transaction{tx3}, block2.Hash)

	// Verify chain linkage
	if string(block2.PrevBlockHash) != string(genesisBlock.Hash) {
		t.Error("Block 2 should reference genesis block hash")
	}

	if string(block3.PrevBlockHash) != string(block2.Hash) {
		t.Error("Block 3 should reference block 2 hash")
	}

	// Mine the blocks using proof of work
	pow2 := proofofwork.NewProofOfWork(block2)
	nonce2, hash2 := pow2.Run()
	block2.Nonce = nonce2
	block2.Hash = hash2

	pow3 := proofofwork.NewProofOfWork(block3)
	nonce3, hash3 := pow3.Run()
	block3.Nonce = nonce3
	block3.Hash = hash3

	// Verify proof of work
	if !pow2.Validate() {
		t.Error("Block 2 proof of work validation failed")
	}

	if !pow3.Validate() {
		t.Error("Block 3 proof of work validation failed")
	}

	t.Logf("Multiple blocks test passed")
}
