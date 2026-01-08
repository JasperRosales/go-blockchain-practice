package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"

	"github.com/aspertheghost/blockchain/internal/block"
	"github.com/aspertheghost/blockchain/internal/proofofwork"
	"github.com/aspertheghost/blockchain/internal/transaction"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) {
	var lastHash []byte

	// Validate transactions
	if len(transactions) == 0 {
		log.Panic("ERROR: No transactions to mine")
	}

	for i, tx := range transactions {
		if tx == nil {
			log.Panic(fmt.Sprintf("ERROR: Transaction at index %d is nil", i))
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return fmt.Errorf("blocks bucket not found")
		}
		lastHash = b.Get([]byte("l"))
		if lastHash == nil {
			return fmt.Errorf("last block not found")
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := block.NewBlock(transactions, lastHash)

	// Run proof of work to mine the block
	pow := proofofwork.NewProofOfWork(newBlock)
	nonce, hash := pow.Run()
	newBlock.Nonce = nonce
	newBlock.Hash = hash

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTransactions(address string) []transaction.Transaction {
	var unspentTXs []transaction.Transaction

	// First pass: collect all outputs that are referenced by any input
	// (these outputs are spent)
	referencedByInput := make(map[string][]int)
	bci := bc.Iterator()

	for {
		b := bci.Next()

		for _, tx := range b.Transactions {
			for _, in := range tx.Vin {
				if in.CanUnlockOutputWith(address) {
					inTxID := hex.EncodeToString(in.Txid)
					referencedByInput[inTxID] = append(referencedByInput[inTxID], in.Vout)
				}
			}
		}

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}

	// Second pass: collect transactions with unspent outputs
	bci = bc.Iterator()

	for {
		b := bci.Next()

		for _, tx := range b.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was this output referenced as an input in any transaction?
				if referencedByInput[txID] != nil {
					for _, spentOut := range referencedByInput[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
		}

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []transaction.TXOutput {
	var UTXOs []transaction.TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, transaction.TXOutputMap) {
	unspentOutputs := make(transaction.TXOutputMap)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *block.Block {
	var blk *block.Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return fmt.Errorf("blocks bucket not found")
		}
		encodedBlock := b.Get(i.currentHash)
		if encodedBlock == nil {
			return fmt.Errorf("block not found")
		}
		blk = block.DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = blk.PrevBlockHash

	return blk
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// DBExists returns true if the blockchain database exists
func DBExists() bool {
	return dbExists()
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			fmt.Println("No existing blockchain found. Please create one first using: ./blockchain createblockchain -address <YOUR_ADDRESS>")
			os.Exit(1)
		}
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// Close closes the database
func (bc *Blockchain) Close() {
	bc.db.Close()
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := transaction.NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := block.NewGenesisBlock(cbtx)

		pow := proofofwork.NewProofOfWork(genesis)
		nonce, hash := pow.Run()

		genesis.Nonce = nonce
		genesis.Hash = hash

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}
