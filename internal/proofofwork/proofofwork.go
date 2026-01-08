package proofofwork

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"

	"github.com/aspertheghost/blockchain/internal/block"
	"github.com/aspertheghost/blockchain/internal/utils"
)

var (
	maxNonce = math.MaxInt64
)

// TargetBits is the difficulty target (lower = easier to mine)
// 16 is a good balance for testing - allows genesis block creation in reasonable time
const targetBits = 16

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *block.Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *block.Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			utils.IntToHex(pow.block.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining genesis block (difficulty: %d bits)...\n", targetBits)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)

		// Print progress every 10000 nonces
		if nonce%10000 == 0 {
			fmt.Printf("\rProgress: nonce=%d, hash=%x", nonce, hash[:12])
		}

		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Printf("\rMining complete! nonce=%d, hash=%x    \n\n", nonce, hash[:])

	return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
